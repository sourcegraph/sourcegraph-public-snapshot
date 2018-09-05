package search

import "regexp/syntax"

// PatternInfo is the struct used by vscode pass on search queries. Keep it in
// sync with pkg/searcher/protocol.PatternInfo.
type PatternInfo struct {
	Pattern         string
	IsRegExp        bool
	IsWordMatch     bool
	IsCaseSensitive bool
	FileMatchLimit  int32

	// We do not support IsMultiline
	//IsMultiline     bool
	IncludePattern  string
	IncludePatterns []string
	ExcludePattern  string

	PathPatternsAreRegExps       bool
	PathPatternsAreCaseSensitive bool

	PatternMatchesContent bool
	PatternMatchesPath    bool
}

func (p *PatternInfo) IsEmpty() bool {
	return p.Pattern == "" && p.ExcludePattern == "" && len(p.IncludePatterns) == 0 && p.IncludePattern == ""
}

// Validate returns a non-nil error if PatternInfo is not valid.
func (p *PatternInfo) Validate() error {
	if p.IsRegExp {
		if _, err := syntax.Parse(p.Pattern, syntax.Perl); err != nil {
			return err
		}
	}

	if p.PathPatternsAreRegExps {
		if p.IncludePattern != "" {
			if _, err := syntax.Parse(p.IncludePattern, syntax.Perl); err != nil {
				return err
			}
		}
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

// Args
type Args struct {
	query *PatternInfo
	//repos []*repositoryRevisions
}
