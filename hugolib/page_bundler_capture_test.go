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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPageBundlerCapture(t *testing.T) {
	t.Parallel()

	assert := require.New(t)
	_, fs := newTestCfg()

	// Bundle
	writeSource(t, fs, filepath.Join("base", "_index.md"), "content")
	writeSource(t, fs, filepath.Join("base", "_1.md"), "content")

	// Singles
	writeSource(t, fs, filepath.Join("base", "images", "hugo-logo.png"), "content")
	writeSource(t, fs, filepath.Join("base", "a", "2.md"), "content")
	writeSource(t, fs, filepath.Join("base", "a", "1.md"), "content")

	// Bundle
	writeSource(t, fs, filepath.Join("base", "b", "index.md"), "content")
	writeSource(t, fs, filepath.Join("base", "b", "1.md"), "content")
	writeSource(t, fs, filepath.Join("base", "b", "2.md"), "content")
	writeSource(t, fs, filepath.Join("base", "b", "c", "logo.png"), "content")

	c := newCapturer(fs.Source, "base")

	assert.NoError(c.capture())

	assert.Len(c.singles, 3)
	assert.Len(c.bundles, 2)

	assert.Len(c.getBundleDirByOwner(filepath.FromSlash("_index.md")).resources, 1)
	assert.Len(c.getBundleDirByOwner(filepath.FromSlash("b/index.md")).resources, 3)
}

func BenchmarkPageBundlerCapture(b *testing.B) {
	capturers := make([]*capturer, b.N)

	for i := 0; i < b.N; i++ {
		_, fs := newTestCfg()
		base := fmt.Sprintf("base%d", i)
		for j := 1; j <= 5; j++ {
			js := fmt.Sprintf("j%d", j)
			writeSource(b, fs, filepath.Join(base, js, "index.md"), "content")
			writeSource(b, fs, filepath.Join(base, js, "logo1.png"), "content")
			writeSource(b, fs, filepath.Join(base, js, "sub", "logo2.png"), "content")
			writeSource(b, fs, filepath.Join(base, js, "section", "_index.md"), "content")
			writeSource(b, fs, filepath.Join(base, js, "section", "logo.png"), "content")
			writeSource(b, fs, filepath.Join(base, js, "section", "sub", "logo.png"), "content")

			for k := 1; k <= 5; k++ {
				ks := fmt.Sprintf("k%d", k)
				writeSource(b, fs, filepath.Join(base, js, ks, "logo1.png"), "content")
				writeSource(b, fs, filepath.Join(base, js, "section", ks, "logo.png"), "content")
			}
		}

		capturers[i] = newCapturer(fs.Source, base)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := capturers[i].capture()
		if err != nil {
			b.Fatal(err)
		}
	}
}
