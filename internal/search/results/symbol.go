package results

import (
	"github.com/sourcegraph/sourcegraph/internal/gituri"
	"github.com/sourcegraph/sourcegraph/internal/symbols/protocol"
)

// SearchSymbolResult is a result from symbol search.
type SearchSymbolResult struct {
	Symbol  protocol.Symbol
	BaseURI *gituri.URI
	Lang    string
}

func (s *SearchSymbolResult) URI() *gituri.URI {
	return s.BaseURI.WithFilePath(s.Symbol.Path)
}
