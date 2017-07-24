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

// TODO(bep) bundles should really try to get this and more in separate package(s)

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gohugoio/hugo/helpers"
	"github.com/spf13/afero"
)

type bundleDirType int

const (
	bundleRegular bundleDirType = iota
	bundleLeaf
	bundleBranch
)

func identifyBundleDir(name string) bundleDirType {
	if strings.HasPrefix(name, "_index.") {
		return bundleBranch
	}

	if strings.HasPrefix(name, "index.") {
		return bundleLeaf
	}

	return bundleRegular
}

func (b bundleDirType) String() string {
	switch b {
	case bundleRegular:
		return "Not a bundle"
	case bundleLeaf:
		return "Regular bundle"
	case bundleBranch:
		return "Branch bundle"
	}
	return ""
}

type bundleDir struct {
	owner     string
	resources []string
}

func newBundleDir(owner string, resources []string) bundleDir {
	return bundleDir{owner: owner, resources: resources}
}

// TODO(bep) bundle benchmark this capture
type capturer struct {
	// To prevent symbolic link cycles: Visit same folder only once.
	seen map[string]bool

	bundles []bundleDir
	singles []string

	fs      afero.Fs
	baseDir string
}

func (c *capturer) getBundleDirByOwner(owner string) bundleDir {
	for _, b := range c.bundles {
		if b.owner == owner {
			return b
		}
	}
	return bundleDir{}
}

func newCapturer(fs afero.Fs, baseDir string) *capturer {
	if !strings.HasSuffix(baseDir, helpers.FilePathSeparator) {
		baseDir += helpers.FilePathSeparator
	}
	return &capturer{fs: fs, baseDir: baseDir, seen: make(map[string]bool)}
}

func (c *capturer) capture() error {
	filenames, err := c.handleDir(c.baseDir)
	if err != nil {
		return err
	}
	c.singles = append(c.singles, filenames...)

	return nil

}

// TODO(bep) bundle test symbolic link
func (c *capturer) handleDir(dirname string) ([]string, error) {
	if c.seen[dirname] {
		return nil, nil
	}
	c.seen[dirname] = true

	files, err := c.readDir(dirname)
	if err != nil {
		return nil, err
	}

	bundleType := bundleRegular
	var (
		bundleFiles []string
		nestedFiles []string
		bundleOwner string
	)

	for _, f := range files {
		filename := filepath.Join(dirname, f)
		fi, err := c.fs.Stat(filename)
		if err != nil {
			// It got deleted in the meantime.
			if !os.IsNotExist(err) {
				return nil, err
			}
			continue
		}

		filenameRelativeToBaseDir := strings.TrimPrefix(filename, c.baseDir)

		if fi.IsDir() {
			filenames, err := c.handleDir(filename)
			if err != nil {
				return nil, err
			}
			nestedFiles = append(nestedFiles, filenames...)
		} else {
			currentIsOwner := false
			if bundleType == bundleRegular {
				bundleType = identifyBundleDir(f)
				currentIsOwner = bundleType != bundleRegular
			}

			if currentIsOwner {
				bundleOwner = filenameRelativeToBaseDir
			} else {
				bundleFiles = append(bundleFiles, filenameRelativeToBaseDir)
			}
		}
	}

	switch bundleType {
	case bundleLeaf:
		// All nested non-bundles are part of this bundle.
		bundleFiles = append(bundleFiles, nestedFiles...)
		c.bundles = append(c.bundles, newBundleDir(bundleOwner, bundleFiles))
	case bundleBranch:
		// All files in the current directory is part of this bundle.
		// Trying to include sub folders in these bundles are filled with ambiguity.
		c.bundles = append(c.bundles, newBundleDir(bundleOwner, bundleFiles))
		c.singles = append(c.singles, nestedFiles...)
	default:
		// Let the parent decide what to do.
		bundleFiles = append(bundleFiles, nestedFiles...)

		return bundleFiles, nil
	}

	return nil, nil
}

// Avoid sorting and FileInfo creation, we don't need that.
func (c *capturer) readDir(dirname string) ([]string, error) {
	dir, err := c.fs.Open(dirname)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	return dir.Readdirnames(-1)
}
