// Copyright 2017 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitindex

import "github.com/grafana/regexp"

// Filter is a include/exclude filter to be used for repo names.
type Filter struct {
	inc, exc *regexp.Regexp
}

// Include returns true if the name passes the filter.
func (f *Filter) Include(name string) bool {
	if f.inc != nil {
		if !f.inc.MatchString(name) {
			return false
		}
	}
	if f.exc != nil {
		if f.exc.MatchString(name) {
			return false
		}
	}
	return true
}

// NewFilter creates a new filter.
func NewFilter(includeRegex, excludeRegex string) (*Filter, error) {
	f := &Filter{}
	var err error
	if includeRegex != "" {
		f.inc, err = regexp.Compile(includeRegex)
		if err != nil {
			return nil, err
		}
	}
	if excludeRegex != "" {
		f.exc, err = regexp.Compile(excludeRegex)
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}
