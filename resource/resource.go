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

package resource

type Type string

// Resources represents a slice of resources, which can be a mix of different types.
// I.e. both pages and images etc.
type Resources []Resource

// Resource represents a linkable resource, i.e. a content page, image etc.
type Resource interface {
	Permalink() string
	RelPermalink() string
	ResourceType() Type
}

// GenericResource represents a generic linkable resource.
type GenericResource struct {
	relPermalink string
	permalink    string
	tp           Type
}

func (l GenericResource) Permalink() string {
	return l.permalink
}

func (l GenericResource) RelPermalink() string {
	return l.relPermalink
}
func (l GenericResource) ResourceType() Type {
	return l.tp
}

func NewGenericResource(relPermalink, permalink string, tp Type) GenericResource {
	return GenericResource{relPermalink: relPermalink, permalink: permalink, tp: tp}
}

var (
	_ Resource = (*GenericResource)(nil)
)
