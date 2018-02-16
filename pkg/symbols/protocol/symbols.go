package protocol

import (
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// SearchArgs are the arguments to perform a search on the symbols service.
type SearchArgs struct {
	// Repo is the repository URI to search in.
	Repo api.RepoURI `json:"repo"`

	// CommitID is the commit to search in.
	CommitID api.CommitID `json:"commitID"`

	// Query is the search query.
	Query string

	// First indicates that only the first n symbols should be returned.
	First int
}

// SearchResult is the result of a search on the symbols service.
type SearchResult struct {
	Symbols []Symbol // code symbols
}

// Symbol is a code symbol.
type Symbol struct {
	Name       string
	Path       string
	Line       int
	Kind       string
	Language   string
	Parent     string
	ParentKind string
	Signature  string
	Pattern    string

	FileLimited bool
}
