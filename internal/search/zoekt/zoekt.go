package zoekt

import (
	"regexp/syntax"

	zoektquery "github.com/google/zoekt/query"
)

func FileRe(pattern string, queryIsCaseSensitive bool) (zoektquery.Q, error) {
	return ParseRe(pattern, true, false, queryIsCaseSensitive)
}

func noOpAnyChar(re *syntax.Regexp) {
	if re.Op == syntax.OpAnyChar {
		re.Op = syntax.OpAnyCharNotNL
	}
	for _, s := range re.Sub {
		noOpAnyChar(s)
	}
}

func ParseRe(pattern string, filenameOnly bool, contentOnly bool, queryIsCaseSensitive bool) (zoektquery.Q, error) {
	// these are the flags used by zoekt, which differ to searcher.
	re, err := syntax.Parse(pattern, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		return nil, err
	}
	noOpAnyChar(re)
	// zoekt decides to use its literal optimization at the query parser
	// level, so we check if our regex can just be a literal.
	if re.Op == syntax.OpLiteral {
		return &zoektquery.Substring{
			Pattern:       string(re.Rune),
			CaseSensitive: queryIsCaseSensitive,
			Content:       contentOnly,
			FileName:      filenameOnly,
		}, nil
	}
	return &zoektquery.Regexp{
		Regexp:        re,
		CaseSensitive: queryIsCaseSensitive,
		Content:       contentOnly,
		FileName:      filenameOnly,
	}, nil
}
