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
	"testing"

	"github.com/gohugoio/hugo/deps"

	"github.com/stretchr/testify/require"
)

func TestPageBundlerSite(t *testing.T) {
	t.Parallel()

	assert := require.New(t)
	cfg, fs := newTestBundleSources(t)

	s := buildSingleSite(t, deps.DepsCfg{Fs: fs, Cfg: cfg}, BuildCfg{})

	// Singles (2), Below home (1), Bundle (1)
	assert.Len(s.RegularPages, 4)

}
