// Package search provides high level search structures and logic.
package search

import (
	"regexp/syntax"
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

	Languages []string
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
