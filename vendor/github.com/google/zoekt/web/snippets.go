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

package web

import (
	"bytes"
	"html/template"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/zoekt"
)

func (s *Server) formatResults(result *zoekt.SearchResult, query string, localPrint bool) ([]*FileMatch, error) {
	var fmatches []*FileMatch

	templateMap := map[string]*template.Template{}
	fragmentMap := map[string]*template.Template{}
	if !localPrint {
		for repo, str := range result.RepoURLs {
			templateMap[repo] = s.getTemplate(str)
		}
		for repo, str := range result.LineFragments {
			fragmentMap[repo] = s.getTemplate(str)
		}
	}
	getFragment := func(repo string, linenum int) string {
		if localPrint {
			return "#l" + strconv.Itoa(linenum)
		}
		if tpl := fragmentMap[repo]; tpl != nil {
			var buf bytes.Buffer
			if err := tpl.Execute(&buf, map[string]string{
				"LineNumber": strconv.Itoa(linenum),
			}); err != nil {
				log.Printf("fragment template: %v", err)
				return ""
			}
			return buf.String()
		}
		return ""
	}
	getURL := func(repo, filename string, branches []string, version string) string {
		if localPrint {
			v := make(url.Values)
			v.Add("r", repo)
			v.Add("f", filename)
			v.Add("q", query)
			if len(branches) > 0 {
				v.Add("b", branches[0])
			}
			return "print?" + v.Encode()
		}

		if tpl := templateMap[repo]; tpl != nil {
			var buf bytes.Buffer
			b := ""
			if len(branches) > 0 {
				b = branches[0]
			}
			err := tpl.Execute(&buf, map[string]string{
				"Branch":  b,
				"Version": version,
				"Path":    filename,
			})
			if err != nil {
				log.Printf("url template: %v", err)
				return ""
			}
			return buf.String()
		}
		return ""
	}

	// hash => result-id
	seenFiles := map[string]string{}
	for _, f := range result.Files {
		fMatch := FileMatch{
			FileName: f.FileName,
			Repo:     f.Repository,
			ResultID: f.Repository + ":" + f.FileName,
			Branches: f.Branches,
			Language: f.Language,
		}

		if dup, ok := seenFiles[string(f.Checksum)]; ok {
			fMatch.DuplicateID = dup
		} else {
			seenFiles[string(f.Checksum)] = fMatch.ResultID
		}

		if f.SubRepositoryName != "" {
			fn := strings.TrimPrefix(fMatch.FileName[len(f.SubRepositoryPath):], "/")
			fMatch.URL = getURL(f.SubRepositoryName, fn, f.Branches, f.Version)
		} else {
			fMatch.URL = getURL(f.Repository, f.FileName, f.Branches, f.Version)
		}

		for _, m := range f.LineMatches {
			fragment := getFragment(f.Repository, m.LineNumber)
			if !strings.HasPrefix(fragment, "#") && !strings.HasPrefix(fragment, ";") {
				// TODO - remove this is backward compatibility glue.
				fragment = "#" + fragment
			}
			md := Match{
				FileName: f.FileName,
				LineNum:  m.LineNumber,
				URL:      fMatch.URL + fragment,
			}

			lastEnd := 0
			line := m.Line
			for i, f := range m.LineFragments {
				l := f.LineOffset
				e := l + f.MatchLength

				frag := Fragment{
					Pre:   string(line[lastEnd:l]),
					Match: string(line[l:e]),
				}
				if i == len(m.LineFragments)-1 {
					frag.Post = string(m.Line[e:])
				}

				md.Fragments = append(md.Fragments, frag)
				lastEnd = e
			}
			fMatch.Matches = append(fMatch.Matches, md)
		}
		fmatches = append(fmatches, &fMatch)
	}
	return fmatches, nil
}
