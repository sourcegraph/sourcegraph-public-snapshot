package result

import (
	"github.com/sourcegraph/sourcegraph/internal/gituri"
)

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

// SearchSymbolResult is a result from symbol search.
type SearchSymbolResult struct {
	Symbol  Symbol
	BaseURI *gituri.URI
	Lang    string
}

func (s *SearchSymbolResult) URI() *gituri.URI {
	return s.BaseURI.WithFilePath(s.Symbol.Path)
}
