package search

import (
	"bytes"
	"regexp/syntax"
	"strings"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/zoekt/query"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search/casetransform"
	"github.com/sourcegraph/sourcegraph/internal/trace"

	zoektquery "github.com/sourcegraph/zoekt/query"
)

type matcher interface {
	AddAttributes(tr trace.Trace)

	MatchesAllContent() bool
	IgnoreCase() bool

	MatchesString(file string) bool
	MatchesFile(fileBuf []byte, limit int) [][]int

	ToZoektQuery(matchContent bool, matchPath bool) (zoektquery.Q, error)
}

// regexMatcher is responsible for finding LineMatches. It is not concurrency
// safe (it reuses buffers for performance).
//
// This code is base on reading the techniques detailed in
// http://blog.burntsushi.net/ripgrep/
//
// The stdlib regexp is pretty powerful and in fact implements many of the
// features in ripgrep. Our implementation gives high performance via pruning
// aggressively which files to consider (non-binary under a limit) and
// optimizing for assuming most lines will not contain a match. The pruning of
// files is done by the
//
// If there is no more low-hanging fruit and perf is not acceptable, we could
// consider using ripgrep directly (modify it to search zip archives).
//
// TODO(keegan) return search statistics
type regexMatcher struct {
	// re is the regexp to match, or nil if empty ("match all files' content").
	re *regexp.Regexp

	// ignoreCase if true means we need to do case insensitive matching.
	ignoreCase bool

	// literalSubstring is used to test if a file is worth considering for
	// matches. literalSubstring is guaranteed to appear in any match found by
	// re. It is the output of the longestLiteral function. It is only set if
	// the regex has an empty LiteralPrefix.
	literalSubstring []byte
}

// compile returns a regexMatcher for matching p.
func compile(p *protocol.PatternInfo) (matcher, error) {
	var (
		re               *regexp.Regexp
		literalSubstring []byte
	)
	if p.Pattern != "" {
		expr := p.Pattern
		if !p.IsRegExp {
			expr = regexp.QuoteMeta(expr)
		}
		if p.IsWordMatch {
			expr = `\b` + expr + `\b`
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
	}

	return &regexMatcher{
		re:               re,
		ignoreCase:       !p.IsCaseSensitive,
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

func (rm *regexMatcher) AddAttributes(tr trace.Trace) {
	if rm.re != nil {
		tr.SetAttributes(attribute.Stringer("re", rm.re))
	}
}

// MatchesAllContent returns whether the regexp pattern will match all content
func (rm *regexMatcher) MatchesAllContent() bool {
	return rm.re == nil
}

func (rm *regexMatcher) IgnoreCase() bool {
	return rm.ignoreCase
}

// MatchesString returns whether the regexp pattern matches s. It is intended to be
// used to match file paths.
func (rm *regexMatcher) MatchesString(s string) bool {
	if rm.re == nil {
		return true
	}
	if rm.ignoreCase {
		s = strings.ToLower(s)
	}
	return rm.re.MatchString(s)
}

// MatchesFile returns a LineMatch for each line that matches rm in reader.
// LimitHit is true if some matches may not have been included in the result.
// NOTE: This is not safe to use concurrently.
func (rm *regexMatcher) MatchesFile(fileBuf []byte, limit int) [][]int {
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
			&zoektquery.Regexp{
				Regexp:        re,
				Content:       true,
				CaseSensitive: !rm.ignoreCase,
			},
			&zoektquery.Regexp{
				Regexp:        re,
				FileName:      true,
				CaseSensitive: !rm.ignoreCase,
			},
		), nil
	}
	return &zoektquery.Regexp{
		Regexp:        re,
		Content:       matchContent,
		FileName:      matchPath,
		CaseSensitive: !rm.ignoreCase,
	}, nil
}
