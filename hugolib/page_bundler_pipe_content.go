// Copyright 2017-present The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hugolib

import (
	"fmt"
	"strings"
)

// TODO(bep) bundle CSS min handler. Document its removal.

type filesOrPage struct {
	bundle   bundleDir
	filename string
	p        *Page
}

type (
	contentItemHandler       func(f *filesOrPage, pages chan<- *Page) (string, error)
	contentItemHandlerCreate func(s *Site) contentItemHandler
	contentItemHandlerRoutes map[string]contentItemHandler
	contentPipeline          struct {
		s      *Site
		routes []contentItemHandlerRoutes
	}

	// contentPipelines is the entry point to content processing. It maps the file
	// extension to a given pipeline.
	// A default (key "*") will be added for the unknown cases.
	contentPipelines struct {
		s     *Site
		pipes map[string]*contentPipeline
	}
)

const (
	defaultContentHandlerRoute = "*"
)

func newContentPipelines(s *Site) *contentPipelines {
	return &contentPipelines{s: s, pipes: make(map[string]*contentPipeline)}
}

func (p *contentPipelines) forAllUnknownFiles() *contentPipeline {
	return p.forExtensions(defaultContentHandlerRoute)
}

func (p *contentPipelines) forExtensions(extensions ...string) *contentPipeline {
	pp := &contentPipeline{s: p.s}
	for _, ext := range extensions {
		p.pipes[ext] = pp
	}
	return pp
}

func newDefaultPipelines(s *Site) *contentPipelines {
	var (
		h     *contentHandlers
		pipes = newContentPipelines(s)
	)

	// The old solution read the sources in a `MetaHandler` picked by its extension in its `Read` method.
	// Then the pages from the first step are sent to a new conversion handler chosen by either the `Markup` variable
	// or the file extension, and the files, i.e. the result from the first step that did not result in a page,
	// are sent into a file handler conversion chosen by its file extension.
	// A mouthful. The main distinction here is the `basicFileHandler` (`mmark`, `html` etc.)
	// that tries to create a `Page`, and `basicFileHandler` (`css`, `unknown` files etc.)
	// where the `Read` method just read into a (lazy) `source.File` object and stored in `Site.Files`.
	//
	// The motivation behind all of this is to adapt it to a concept of bundles (i.e. we cannot think in single files terms only),
	// but first we got to get the existing flow working.
	//
	// This is the outline of the handler chain:
	// CSS: reader|minimize|reader to destination
	// Unknown: reader|reader to destination
	// MMark: reader|create page|page handler (by Markup or file ext)
	// HTML: reader|create page|page handler
	//
	// The below translates to:
	// 1. Creates a Page and reads front matter.
	// 2.1 the content page handler for all but html/htm files (from Markup front matter or file extension).
	// 2.2 the content HTML page handler for all html/htm files.
	pipes.forExtensions(
		"html", "htm",
		"mdown", "markdown", "md",
		"asciidoc", "adoc", "ad",
		"rest", "rst",
		"mmark",
		"org").
		do(h.contentPageFirstStep).
		do(h.contentPageHandle).
		or(h.contentHTMLPageHandle, "html", "htm")

	// Finally we set up a default pipeline for all the other files (jpeg images etc.),
	// that is just copied to destination.
	pipes.forAllUnknownFiles().do(h.copyContentFileToDestination)

	return pipes
}

func (cp contentPipelines) getMatchingOrDefault(ext string) *contentPipeline {
	ext = strings.TrimPrefix(ext, ".")
	if handler, found := cp.pipes[ext]; found {
		return handler
	}
	return cp.pipes[defaultContentHandlerRoute]
}

func (p *contentPipeline) do(h contentItemHandlerCreate) *contentPipeline {
	routes := make(contentItemHandlerRoutes)
	routes.addDefault(h(p.s))
	p.routes = append(p.routes, routes)
	return p
}

func (p *contentPipeline) or(h contentItemHandlerCreate, matchers ...string) *contentPipeline {
	route := p.routes[len(p.routes)-1]
	handler := h(p.s)
	for _, matcher := range matchers {
		route.add(matcher, handler)
	}

	return p
}

func (r contentItemHandlerRoutes) add(id string, h contentItemHandler) {
	if _, found := r[id]; found {
		panic(fmt.Sprintf("Handler with id %q already registered", id))
	}
	r[id] = h
}

func (r contentItemHandlerRoutes) addDefault(h contentItemHandler) {
	r.add(defaultContentHandlerRoute, h)
}

func (r contentItemHandlerRoutes) get(id string) contentItemHandler {
	if route, ok := r[id]; ok {
		return route
	}
	return r.getDefault()
}

func (r contentItemHandlerRoutes) getDefault() contentItemHandler {
	return r[defaultContentHandlerRoute]
}
