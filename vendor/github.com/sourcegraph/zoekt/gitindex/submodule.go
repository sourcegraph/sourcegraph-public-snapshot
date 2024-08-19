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

package gitindex

import (
	"bytes"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/format/config"
)

// SubmoduleEntry represent one entry in a .gitmodules file
type SubmoduleEntry struct {
	Path   string
	URL    string
	Branch string
}

// ParseGitModules parses the contents of a .gitmodules file.
func ParseGitModules(content []byte) (map[string]*SubmoduleEntry, error) {
	buf := bytes.NewBuffer(content)

	// Handle the possibility that .gitmodules has a UTF-8 BOM, which would
	// otherwise break the scanner.
	// https://stackoverflow.com/a/21375405
	skipIfPrefix(buf, []byte("\uFEFF"))

	dec := config.NewDecoder(buf)
	cfg := &config.Config{}

	if err := dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("error decoding content %s: %w", string(content), err)
	}

	result := map[string]*SubmoduleEntry{}
	for _, s := range cfg.Sections {
		if s.Name != "submodule" {
			continue
		}

		for _, ss := range s.Subsections {
			name := ss.Name
			e := &SubmoduleEntry{}
			for _, o := range ss.Options {
				switch o.Key {
				case "branch":
					e.Branch = o.Value
				case "path":
					e.Path = o.Value
				case "url":
					e.URL = o.Value
				}
			}

			result[name] = e
		}
	}

	return result, nil
}

// skipIfPrefix will detect if the unread portion of buf starts with
// prefix. If it does, it will read over those bytes.
func skipIfPrefix(buf *bytes.Buffer, prefix []byte) {
	if bytes.HasPrefix(buf.Bytes(), prefix) {
		buf.Next(len(prefix))
	}
}
