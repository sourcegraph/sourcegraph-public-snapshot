// Package search provides high level search structures and logic.
package search

import (
	"regexp/syntax"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
)

// PatternInfo is the struct used by vscode pass on search queries. Keep it in
// sync with pkg/searcher/protocol.PatternInfo.
type PatternInfo struct {
	Pattern         string
	IsRegExp        bool
	IsStructuralPat bool
	CombyRule       string
	IsWordMatch     bool
	IsCaseSensitive bool
	FileMatchLimit  int32

	// We do not support IsMultiline
	// IsMultiline     bool
	IncludePatterns []string
	ExcludePattern  string

	FilePatternsReposMustInclude []string
	FilePatternsReposMustExclude []string

	PathPatternsAreRegExps       bool
	PathPatternsAreCaseSensitive bool

	PatternMatchesContent bool
	PatternMatchesPath    bool
}

func (p *PatternInfo) IsEmpty() bool {
	return p.Pattern == "" && p.ExcludePattern == "" && len(p.IncludePatterns) == 0
}

// Validate returns a non-nil error if PatternInfo is not valid.
func (p *PatternInfo) Validate() error {
	if p.IsRegExp {
		if _, err := syntax.Parse(p.Pattern, syntax.Perl); err != nil {
			return err
		}
	}

	if p.PathPatternsAreRegExps {
		if p.ExcludePattern != "" {
			if _, err := syntax.Parse(p.ExcludePattern, syntax.Perl); err != nil {
				return err
			}
		}
		for _, expr := range p.IncludePatterns {
			if _, err := syntax.Parse(expr, syntax.Perl); err != nil {
				return err
			}
		}
	}

	return nil
}

// Args are the arguments passed to a search backend. It contains the Pattern
// to search for, as well as the hydrated list of repository revisions to
// search.
type Args struct {
	Pattern *PatternInfo
	Repos   []*RepositoryRevisions

	// Query is the parsed query from the user. You should be using Pattern
	// instead, but Query is useful for checking extra fields that are set and
	// ignored by Pattern, such as index:no
	Query *query.Query

	// UseFullDeadline indicates that the search should try do as much work as
	// it can within context.Deadline. If false the search should try and be
	// as fast as possible, even if a "slow" deadline is set.
	//
	// For example searcher will wait to full its archive cache for a
	// repository if this field is true. Another example is we set this field
	// to true if the user requests a specific timeout or maximum result size.
	UseFullDeadline bool

	Zoekt        *searchbackend.Zoekt
	SearcherURLs *endpoint.Map
}
