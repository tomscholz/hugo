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
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/gohugoio/hugo/helpers"
)

// A contentSourceMap contains a files map from a walk of the `content` directory.
type contentSourceMap struct {
	s *Site

	c *capturer

	contentFileHandlers *contentPipelines
}

func newBundler(s *Site) *contentSourceMap {
	handlers := newDefaultPipelines(s)

	return &contentSourceMap{
		s:                   s,
		c:                   newCapturer(s.Fs.Source, s.absContentDir()),
		contentFileHandlers: handlers}
}

func (b *contentSourceMap) captureFiles() error {
	return b.c.capture()
}

func (b *contentSourceMap) processContentBundles() error {

	s := b.s

	// The input file bundles.
	fileBundlesChan := make(chan bundleDir)

	// The input file singles.
	fileSinglesChan := make(chan string)

	// The output file batches.
	pagesChan := make(chan *Page)

	numWorkers := getGoMaxProcs() * 4

	var (
		ctx, cancel = context.WithCancel(context.Background())

		filesBundlesProcessors, _ = errgroup.WithContext(ctx)
		filesSinglesProcessors, _ = errgroup.WithContext(ctx)
		pagesCollector, _         = errgroup.WithContext(ctx)
	)
	defer cancel()

	// TODO(bep) bundles tune numWorkers
	for i := 0; i < numWorkers; i++ {
		filesBundlesProcessors.Go(func() error {
			for bundle := range fileBundlesChan {
				// TODO(bep) bundles
				err := b.readAndConvertContentFile(bundle.owner, pagesChan)
				if err != nil {
					return err
				}
				for _, filename := range bundle.resources {
					err := b.readAndConvertContentFile(filename, pagesChan)
					if err != nil {
						return err
					}
				}

			}
			return nil
		})

		filesSinglesProcessors.Go(func() error {
			for filename := range fileSinglesChan {
				// TODO(bep) bundles
				err := b.readAndConvertContentFile(filename, pagesChan)
				if err != nil {
					return err
				}

			}
			return nil
		})
	}

	// There can be only one page collector.
	pagesCollector.Go(func() error {
		for p := range pagesChan {
			s.addPage(p)
		}
		s.rawAllPages.Sort()
		return nil
	})

	for _, filename := range b.c.singles {
		fileSinglesChan <- filename
	}

	close(fileSinglesChan)

	for _, bundleDir := range b.c.bundles {
		fileBundlesChan <- bundleDir
	}

	close(fileBundlesChan)

	f1Err := filesBundlesProcessors.Wait()
	f2Err := filesSinglesProcessors.Wait()

	cancel()

	close(pagesChan)
	pErr := pagesCollector.Wait()

	if f1Err != nil {
		return f1Err
	}
	if f2Err != nil {
		return f2Err
	}
	if pErr != nil {
		return pErr
	}

	return nil
}

func (b *contentSourceMap) readAndConvertContentFile(filename string, pages chan<- *Page) error {

	fop := &filesOrPage{filename: filename}

	ext := helpers.Ext(filename)

	pipeline := b.contentFileHandlers.getMatchingOrDefault(ext)

	if pipeline == nil {
		return nil
	}

	var (
		nextRoute = defaultContentHandlerRoute
		err       error
	)

	for _, route := range pipeline.routes {
		handle := route.get(nextRoute)
		if nextRoute, err = handle(fop, pages); err != nil {
			return err
		}
		if nextRoute == "" {
			// End of Pipe
			break
		}
	}

	return nil
}
