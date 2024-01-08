package search

import (
	"bytes"
	"fmt"
	"regexp/syntax"
	"strings"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/zoekt/query"
	zoektquery "github.com/sourcegraph/zoekt/query"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search/casetransform"
)

type matcher interface {
	// MatchesString returns whether the string matches
	MatchesString(s string) bool

	// MatchesFile returns whether the file matches, plus a LineMatch for each line that matches
	MatchesFile(fileBuf []byte, limit int) (match bool, matches [][]int)

	// ToZoektQuery returns a zoekt query representing the same rules as as this matcher
	ToZoektQuery(matchContent bool, matchPath bool) (zoektquery.Q, error)

	// String returns a simple string representation of the matcher
	String() string
}

// compilePattern returns a matcher for matching the pattern info
func compilePattern(p *protocol.PatternInfo) (matcher, error) {
	var (
		re               *regexp.Regexp
		literalSubstring []byte
	)

	if p.Pattern == "" {
		return &allMatcher{}, nil
	}

	expr := p.Pattern
	if !p.IsRegExp {
		expr = regexp.QuoteMeta(expr)
	}

	if p.IsRegExp {
		// We don't do the search line by line, therefore we want the
		// regex engine to consider newlines for anchors (^$).
		expr = "(?m:" + expr + ")"
	}

	// Transforms on the parsed regex
	{
		re, err := syntax.Parse(expr, syntax.Perl)
		if err != nil {
			return nil, err
		}

		if !p.IsCaseSensitive {
			// We don't just use (?i) because regexp library doesn't seem
			// to contain good optimizations for case insensitive
			// search. Instead we lowercase the input and pattern.
			casetransform.LowerRegexpASCII(re)
		}

		// OptimizeRegexp currently only converts capture groups into
		// non-capture groups (faster for stdlib regexp to execute).
		re = query.OptimizeRegexp(re, syntax.Perl)

		expr = re.String()
	}

	var err error
	re, err = regexp.Compile(expr)
	if err != nil {
		return nil, err
	}

	// Only use literalSubstring optimization if the regex engine doesn't
	// have a prefix to use.
	if pre, _ := re.LiteralPrefix(); pre == "" {
		ast, err := syntax.Parse(expr, syntax.Perl)
		if err != nil {
			return nil, err
		}
		ast = ast.Simplify()
		literalSubstring = []byte(longestLiteral(ast))
	}

	return &regexMatcher{
		re:               re,
		ignoreCase:       !p.IsCaseSensitive,
		isNegated:        p.IsNegated,
		literalSubstring: literalSubstring,
	}, nil
}

// longestLiteral finds the longest substring that is guaranteed to appear in
// a match of re.
//
// Note: There may be a longer substring that is guaranteed to appear. For
// example we do not find the longest common substring in alternating
// group. Nor do we handle concatting simple capturing groups.
func longestLiteral(re *syntax.Regexp) string {
	switch re.Op {
	case syntax.OpLiteral:
		return string(re.Rune)
	case syntax.OpCapture, syntax.OpPlus:
		return longestLiteral(re.Sub[0])
	case syntax.OpRepeat:
		if re.Min >= 1 {
			return longestLiteral(re.Sub[0])
		}
	case syntax.OpConcat:
		longest := ""
		for _, sub := range re.Sub {
			l := longestLiteral(sub)
			if len(l) > len(longest) {
				longest = l
			}
		}
		return longest
	}
	return ""
}

type allMatcher struct{}

func (a allMatcher) MatchesString(_ string) bool {
	return true
}

func (a allMatcher) MatchesFile(_ []byte, _ int) (match bool, matches [][]int) {
	return true, nil
}

func (a allMatcher) ToZoektQuery(_ bool, _ bool) (zoektquery.Q, error) {
	return &zoektquery.Const{Value: true}, nil
}

func (a allMatcher) String() string {
	return "all"
}

type regexMatcher struct {
	// re is the regexp to match, should never be nil
	re *regexp.Regexp

	// ignoreCase if true means we need to do case insensitive matching.
	ignoreCase bool

	// isNegated indicates whether matches on the pattern should be negated (representing a 'NOT' in the query)
	isNegated bool

	// literalSubstring is used to test if a file is worth considering for
	// matches. literalSubstring is guaranteed to appear in any match found by
	// re. It is the output of the longestLiteral function. It is only set if
	// the regex has an empty LiteralPrefix.
	literalSubstring []byte
}

func (rm *regexMatcher) String() string {
	return fmt.Sprintf("re: %q", rm.re)
}

func (rm *regexMatcher) MatchesString(s string) bool {
	matches := rm.matchesString(s)
	return matches == !rm.isNegated
}

func (rm *regexMatcher) matchesString(s string) bool {
	if rm.ignoreCase {
		s = strings.ToLower(s)
	}
	return rm.re.MatchString(s)
}

func (rm *regexMatcher) MatchesFile(fileBuf []byte, limit int) (bool, [][]int) {
	matches := rm.matchesFile(fileBuf, limit)
	match := len(matches) > 0
	return match == !rm.isNegated, matches
}

func (rm *regexMatcher) matchesFile(fileBuf []byte, limit int) [][]int {
	// Most files will not have a match and we bound the number of matched
	// files we return. So we can avoid the overhead of parsing out new lines
	// and repeatedly running the regex engine by running a single match over
	// the whole file. This does mean we duplicate work when actually
	// searching for results. We use the same approach when we search
	// per-line. Additionally if we have a non-empty literalSubstring, we use
	// that to prune out files since doing bytes.Index is very fast.
	if !bytes.Contains(fileBuf, rm.literalSubstring) {
		return nil
	}

	// find limit+1 matches so we know whether we hit the limit
	return rm.re.FindAllIndex(fileBuf, limit+1)
}

func (rm *regexMatcher) ToZoektQuery(matchContent bool, matchPath bool) (zoektquery.Q, error) {
	re, err := syntax.Parse(rm.re.String(), syntax.Perl)
	if err != nil {
		return nil, err
	}
	re = zoektquery.OptimizeRegexp(re, syntax.Perl)

	if matchContent && matchPath {
		return zoektquery.NewOr(
			rm.negateIfNeeded(
				&zoektquery.Regexp{
					Regexp:        re,
					Content:       true,
					CaseSensitive: !rm.ignoreCase,
				}),
			rm.negateIfNeeded(
				&zoektquery.Regexp{
					Regexp:        re,
					FileName:      true,
					CaseSensitive: !rm.ignoreCase,
				}),
		), nil
	}

	return rm.negateIfNeeded(
		&zoektquery.Regexp{
			Regexp:        re,
			Content:       matchContent,
			FileName:      matchPath,
			CaseSensitive: !rm.ignoreCase,
		}), nil
}

func (rm *regexMatcher) negateIfNeeded(q zoektquery.Q) zoektquery.Q {
	if rm.isNegated {
		return &zoektquery.Not{Child: q}
	}
	return q
}
