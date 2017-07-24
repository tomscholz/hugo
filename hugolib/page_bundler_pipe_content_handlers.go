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

	"github.com/gohugoio/hugo/helpers"
)

type contentHandlers struct{}

func (*contentHandlers) contentPageFirstStep(s *Site) contentItemHandler {
	return func(f *fileOrPage, pages chan<- *Page, files chan<- *file) (string, error) {
		nextRoute := defaultContentHandlerRoute

		p, err := s.NewPage(f.f.filename)
		if err != nil {
			return nextRoute, err
		}

		_, err = p.ReadFrom(f.f)
		if err != nil {
			return nextRoute, err
		}

		if !p.shouldBuild() {
			pages <- p
			return nextRoute, nil
		}

		f.p = p

		if p.Markup != "" {
			nextRoute = p.Markup
		} else {
			nextRoute = p.Ext()
		}

		return nextRoute, nil
	}
}

func (*contentHandlers) contentPageHandle(s *Site) contentItemHandler {
	return func(f *fileOrPage, pages chan<- *Page, files chan<- *file) (string, error) {

		nextRoute := defaultContentHandlerRoute

		if f.p == nil {
			return nextRoute, nil
		}

		p := f.p

		// In a multilanguage setup, we use the first site to
		// do the initial processing.
		// That site may be different than where the page will end up,
		// so we do the assignment here.
		// We should clean up this, but that will have to wait.
		s.assignSiteByLanguage(p)

		// TODO(bep) bundler
		if p.rendered {
			panic(fmt.Sprintf("Page %q already rendered, does not need conversion", p.BaseFileName()))
		}

		// Work on a copy of the raw content from now on.
		p.createWorkContentCopy()

		if err := p.processShortcodes(); err != nil {
			p.s.Log.ERROR.Println(err)
		}

		if s.Cfg.GetBool("enableEmoji") {
			p.workContent = helpers.Emojify(p.workContent)
		}

		p.workContent = p.replaceDivider(p.workContent)
		p.workContent = p.renderContent(p.workContent)

		pages <- p

		return nextRoute, nil
	}
}

func (*contentHandlers) contentHTMLPageHandle(s *Site) contentItemHandler {
	return func(f *fileOrPage, pages chan<- *Page, files chan<- *file) (string, error) {

		nextRoute := defaultContentHandlerRoute

		p := f.p

		// TODO(bep) bundler
		if p.rendered {
			panic(fmt.Sprintf("Page %q already rendered, does not need conversion", p.BaseFileName()))
		}

		p.createWorkContentCopy()

		if err := p.processShortcodes(); err != nil {
			p.s.Log.ERROR.Println(err)
		}

		pages <- p

		return nextRoute, nil
	}
}

// TODO(bep) bundle use the files chan, maybe.
func (*contentHandlers) copyContentFileToDestination(s *Site) contentItemHandler {
	return func(f *fileOrPage, pages chan<- *Page, files chan<- *file) (string, error) {
		// TODO(bep) bundle some kind of EOF
		nextRoute := defaultContentHandlerRoute

		return nextRoute, s.publish(f.f.filename, f.f)

	}
}
