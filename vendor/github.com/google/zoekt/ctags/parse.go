// Copyright 2016 Google Inc. All rights reserved.
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

package ctags

import (
	"fmt"
	"strconv"
	"strings"
)

type Entry struct {
	Sym        string
	Path       string
	Line       int
	Kind       string
	Language   string
	Parent     string
	ParentType string

	FileLimited bool
}

// Parse parses a single line of exuberant "ctags -n" output.
func Parse(in string) (*Entry, error) {
	fields := strings.Split(in, "\t")
	e := Entry{}

	if len(fields) < 3 {
		return nil, fmt.Errorf("too few fields: %q", in)
	}

	e.Sym = fields[0]
	e.Path = fields[1]

	lstr := fields[2]
	if len(lstr) < 2 {
		return nil, fmt.Errorf("got %q for linenum field", lstr)
	}

	l, err := strconv.ParseInt(lstr[:len(lstr)-2], 10, 64)
	if err != nil {
		return nil, err
	}
	e.Line = int(l)
	e.Kind = fields[3]

field:
	for _, f := range fields[3:] {
		if string(f) == "file:" {
			e.FileLimited = true
		}
		for _, p := range []string{"class", "enum"} {
			if strings.HasPrefix(f, p+":") {
				e.Parent = strings.TrimPrefix(f, p+":")
				e.ParentType = p
				continue field
			}
		}
	}
	return &e, nil
}
