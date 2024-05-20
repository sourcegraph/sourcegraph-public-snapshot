package zoektquery

import (
	"regexp/syntax" //nolint:depguard // zoekt requires this pkg

	"github.com/sourcegraph/zoekt/query"
)

const regexpFlags = syntax.ClassNL | syntax.PerlX | syntax.UnicodeGroups

func FileRe(pattern string, queryIsCaseSensitive bool) (query.Q, error) {
	return ParseRe(pattern, true, false, queryIsCaseSensitive)
}

func ParseRe(pattern string, filenameOnly bool, contentOnly bool, queryIsCaseSensitive bool) (query.Q, error) {
	// these are the flags used by zoekt, which differ to searcher.
	re, err := syntax.Parse(pattern, regexpFlags)
	if err != nil {
		return nil, err
	}

	// OptimizeRegexp currently only converts capture groups into non-capture
	// groups (faster for stdlib regexp to execute).
	re = query.OptimizeRegexp(re, regexpFlags)

	// zoekt decides to use its literal optimization at the query parser
	// level, so we check if our regex can just be a literal.
	if re.Op == syntax.OpLiteral {
		return &query.Substring{
			Pattern:       string(re.Rune),
			CaseSensitive: queryIsCaseSensitive,
			Content:       contentOnly,
			FileName:      filenameOnly,
		}, nil
	}
	return &query.Regexp{
		Regexp:        re,
		CaseSensitive: queryIsCaseSensitive,
		Content:       contentOnly,
		FileName:      filenameOnly,
	}, nil
}
