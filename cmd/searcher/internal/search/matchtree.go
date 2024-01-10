package search

import (
	"bytes"
	"fmt"
	"regexp/syntax"
	"sort"
	"strings"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/zoekt/query"
	zoektquery "github.com/sourcegraph/zoekt/query"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search/casetransform"
)

type matchTree interface {
	// MatchesString returns whether the string matches
	MatchesString(s string) bool

	// MatchesFile returns whether the file matches, plus a LineMatch for each line that matches.
	// Note: even if the returned matches slice is empty, match can be true. This can happen if
	// a query matches all content, or if the query is negated.
	MatchesFile(fileBuf []byte, limit int) (match bool, matches [][]int)

	// ToZoektQuery returns a zoekt query representing the same rules as as this matchTree
	ToZoektQuery(matchContent bool, matchPath bool) (zoektquery.Q, error)

	// String returns a simple string representation of the matchTree
	String() string
}

// compilePattern returns a matchTree for matching the pattern info
func compilePattern(p *protocol.PatternInfo) (matchTree, error) {
	var (
		re               *regexp.Regexp
		literalSubstring []byte
	)

	if p.Pattern == "" {
		return &allMatchTree{}, nil
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

	return &regexMatchTree{
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

type allMatchTree struct{}

func (a allMatchTree) MatchesString(_ string) bool {
	return true
}

func (a allMatchTree) MatchesFile(_ []byte, _ int) (match bool, matches [][]int) {
	return true, nil
}

func (a allMatchTree) ToZoektQuery(_ bool, _ bool) (zoektquery.Q, error) {
	return &zoektquery.Const{Value: true}, nil
}

func (a allMatchTree) String() string {
	return "all"
}

type regexMatchTree struct {
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

func (rm *regexMatchTree) String() string {
	return fmt.Sprintf("re: %q", rm.re)
}

func (rm *regexMatchTree) MatchesString(s string) bool {
	matches := rm.matchesString(s)
	return matches == !rm.isNegated
}

func (rm *regexMatchTree) matchesString(s string) bool {
	if rm.ignoreCase {
		s = strings.ToLower(s)
	}
	return rm.re.MatchString(s)
}

func (rm *regexMatchTree) MatchesFile(fileBuf []byte, limit int) (bool, [][]int) {
	matches := rm.matchesFile(fileBuf, limit)
	match := len(matches) > 0
	return match == !rm.isNegated, matches
}

func (rm *regexMatchTree) matchesFile(fileBuf []byte, limit int) [][]int {
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

	return rm.re.FindAllIndex(fileBuf, limit)
}

func (rm *regexMatchTree) ToZoektQuery(matchContent bool, matchPath bool) (zoektquery.Q, error) {
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

func (rm *regexMatchTree) negateIfNeeded(q zoektquery.Q) zoektquery.Q {
	if rm.isNegated {
		return &zoektquery.Not{Child: q}
	}
	return q
}

type andMatchTree struct {
	children []matchTree
}

func (am *andMatchTree) MatchesString(s string) bool {
	for _, m := range am.children {
		if !m.MatchesString(s) {
			return false
		}
	}
	return true
}

func (am *andMatchTree) MatchesFile(fileBuf []byte, limit int) (bool, [][]int) {
	var matches [][]int
	for _, m := range am.children {
		// Pass the full limit to the children instead of tracking how many matches we
		// have left. This is slightly wasteful but keeps the logic simpler.
		childMatch, childMatches := m.MatchesFile(fileBuf, limit)
		if !childMatch {
			return false, nil
		}
		matches = append(matches, childMatches...)
	}

	return true, mergeMatches(matches, limit)
}

func (am *andMatchTree) ToZoektQuery(matchContent bool, matchPath bool) (zoektquery.Q, error) {
	var children []zoektquery.Q
	for _, m := range am.children {
		q, err := m.ToZoektQuery(matchContent, matchPath)
		if err != nil {
			return nil, err
		}
		children = append(children, q)
	}
	return &zoektquery.And{Children: children}, nil
}

func (am *andMatchTree) String() string {
	return fmt.Sprintf("AND (%d children)", len(am.children))
}

type orMatchTree struct {
	children []matchTree
}

func (om *orMatchTree) MatchesString(s string) bool {
	for _, m := range om.children {
		if m.MatchesString(s) {
			return true
		}
	}
	return false
}

func (om *orMatchTree) MatchesFile(fileBuf []byte, limit int) (bool, [][]int) {
	match := false
	var matches [][]int
	for _, m := range om.children {
		// Pass the full limit to the children instead of tracking how many matches we
		// have left. This is slightly wasteful but keeps the logic simpler.
		childMatch, childMatches := m.MatchesFile(fileBuf, limit)
		match = match || childMatch
		matches = append(matches, childMatches...)
	}

	return match, mergeMatches(matches, limit)
}

func (om *orMatchTree) ToZoektQuery(matchContent bool, matchPath bool) (zoektquery.Q, error) {
	var children []zoektquery.Q
	for _, m := range om.children {
		q, err := m.ToZoektQuery(matchContent, matchPath)
		if err != nil {
			return nil, err
		}
		children = append(children, q)
	}
	return &zoektquery.Or{Children: children}, nil
}

func (om *orMatchTree) String() string {
	return fmt.Sprintf("OR (%d children)", len(om.children))
}

// mergeMatches sorts the matched ranges and truncates the list
// to obey limit. Consistent with behavior for diff/ commit search,
// it does not merge or remove overlapping ranges.
func mergeMatches(matches [][]int, limit int) [][]int {
	sort.Sort(matchSlice(matches))
	if len(matches) > limit {
		return matches[:limit]
	}
	return matches
}

type matchSlice [][]int

func (ms matchSlice) Len() int {
	return len(ms)
}

func (ms matchSlice) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}

func (ms matchSlice) Less(i, j int) bool {
	return ms[i][0] < ms[j][0]
}
