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
	"time"

	"github.com/google/zoekt"
)

type LastInput struct {
	Query string
	Num   int

	// If set, focus on the search box.
	AutoFocus bool
}

// Result holds the data provided to the search results template.
type ResultInput struct {
	Last          LastInput
	QueryStr      string
	Query         string
	Stats         zoekt.Stats
	Duration      time.Duration
	FileMatches   []*FileMatch
	SearchOptions string
}

// FileMatch holds the per file data provided to search results template
type FileMatch struct {
	FileName string
	Repo     string
	ResultID string
	Language string
	// If this was a duplicate result, this will contain the file
	// of the first match.
	DuplicateID string

	Branches []string
	Matches  []Match
	URL      string
}

// Match holds the per line data provided to the search results template
type Match struct {
	URL      string
	FileName string
	LineNum  int

	Fragments []Fragment
}

// Fragment holds data of a single contiguous match within in a line
// for the results template.
type Fragment struct {
	Pre   string
	Match string
	Post  string
}

// SearchBoxInput is provided to the SearchBox template.
type SearchBoxInput struct {
	Last    LastInput
	Stats   *zoekt.RepoStats
	Version string
	Uptime  time.Duration
}

// RepoListInput is provided to the RepoList template.
type RepoListInput struct {
	Last  LastInput
	Stats zoekt.RepoStats
	Repos []Repository
}

// Branch holds the metadata for a indexed branch.
type Branch struct {
	Name    string
	Version string
	URL     string
}

// Repository holds the metadata for an indexed repository.
type Repository struct {
	Name      string
	URL       string
	IndexTime time.Time
	Branches  []Branch
	Files     int64

	// Total amount of content bytes.
	Size int64
}

// PrintInput is provided to the server.Print template.
type PrintInput struct {
	Repo, Name string
	Lines      []string
	Last       LastInput
}
