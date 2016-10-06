package vcs

import "time"

type Commit struct {
	ID        CommitID   `json:"ID,omitempty"`
	Author    Signature  `json:"Author"`
	Committer *Signature `json:"Committer,omitempty"`
	Message   string     `json:"Message,omitempty"`
	// Parents are the commit IDs of this commit's parent commits.
	Parents []CommitID `json:"Parents,omitempty"`
}

type Signature struct {
	Name  string    `json:"Name,omitempty"`
	Email string    `json:"Email,omitempty"`
	Date  time.Time `json:"Date"`
}

// A Branch is a VCS branch.
type Branch struct {
	// Name is the name of this branch.
	Name string `json:"Name,omitempty"`
	// Head is the commit ID of this branch's head commit.
	Head CommitID `json:"Head,omitempty"`
	// Commit optionally contains commit information for this branch's head commit.
	// It is populated if IncludeCommit option is set.
	Commit *Commit `json:"Commit,omitempty"`
	// Counts optionally contains the commit counts relative to specified branch.
	Counts *BehindAhead `json:"Counts,omitempty"`
}

// BehindAhead is a set of behind/ahead counts.
type BehindAhead struct {
	Behind uint32 `json:"Behind,omitempty"`
	Ahead  uint32 `json:"Ahead,omitempty"`
}

// BranchesOptions specifies options for the list of branches returned by
// (Repository).Branches.
type BranchesOptions struct {
	// MergedInto will cause the returned list to be restricted to only
	// branches that were merged into this branch name.
	MergedInto string `json:"MergedInto,omitempty" url:",omitempty"`
	// IncludeCommit controls whether complete commit information is included.
	IncludeCommit bool `json:"IncludeCommit,omitempty" url:",omitempty"`
	// BehindAheadBranch specifies a branch name. If set to something other than blank
	// string, then each returned branch will include a behind/ahead commit counts
	// information against the specified base branch. If left blank, then branches will
	// not include that information and their Counts will be nil.
	BehindAheadBranch string `json:"BehindAheadBranch,omitempty" url:",omitempty"`
	// ContainsCommit filters the list of branches to only those that
	// contain a specific commit ID (if set).
	ContainsCommit string `json:"ContainsCommit,omitempty" url:",omitempty"`
}

// A Tag is a VCS tag.
type Tag struct {
	Name     string   `json:"Name,omitempty"`
	CommitID CommitID `json:"CommitID,omitempty"`
}

// SearchOptions specifies options for a repository search.
type SearchOptions struct {
	// the query string
	Query string `json:"Query,omitempty"`
	// currently only FixedQuery ("fixed") is supported
	QueryType string `json:"QueryType,omitempty"`
	// the number of lines before and after each hit to display
	ContextLines int32 `json:"ContextLines,omitempty"`
	// max number of matches to return
	N int32 `json:"N,omitempty"`
	// starting offset for matches (use with N for pagination)
	Offset int32 `json:"Offset,omitempty"`
}

// A SearchResult is a match returned by a search.
type SearchResult struct {
	// File is the file that contains this match.
	File string `json:"File,omitempty"`
	// The byte range [start,end) of the match.
	StartByte uint32 `json:"StartByte,omitempty"`
	EndByte   uint32 `json:"EndByte,omitempty"`
	// The line range [start,end] of the match.
	StartLine uint32 `json:"StartLine,omitempty"`
	EndLine   uint32 `json:"EndLine,omitempty"`
	// Match is the matching portion of the file from [StartByte,
	// EndByte).
	Match []byte `json:"Match,omitempty"`
}

// A Committer is a contributor to a repository.
type Committer struct {
	Name    string `json:"Name,omitempty"`
	Email   string `json:"Email,omitempty"`
	Commits int32  `json:"Commits,omitempty"`
}
