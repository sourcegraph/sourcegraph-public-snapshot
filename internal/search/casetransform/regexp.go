package casetransform

import (
	"regexp/syntax"

	"github.com/grafana/regexp"
)

// Regexp is a light wrapper over *regexp.Regexp that optimizes for case-insensitive search.
//
// Case-insensitive search using *regexp.Regexp and `(?i)` meta tags is quite
// slow. To mitigate the performance cost of case-insensitive search, we
// transform regexp patterns to their lower-case equivalent (LowerRegexpASCII),
// and transform the search content to its lower-case equivalent (BytesToLowerASCII)
// before matching the pattern to the content.
//
// This type encodes the requirements that, if ignoreCase is set:
// 1) The regexp pattern is transformed into its lower-case equivalent
// 2) The content to be searched is transformed into its lower-case equivalent
// 3) A re-usable buffer is passed in to the match methods to encourage buffer re-use
type Regexp struct {
	re         *regexp.Regexp
	ignoreCase bool
}

func CompileRegexp(expr string, ignoreCase bool) (*Regexp, error) {
	if ignoreCase {
		syn, err := syntax.Parse(expr, syntax.Perl)
		if err != nil {
			return nil, err
		}
		LowerRegexpASCII(syn)
		expr = syn.String()
	}

	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &Regexp{
		re:         re,
		ignoreCase: ignoreCase,
	}, nil
}

func (r *Regexp) FindAllIndex(b []byte, n int, lowerBuf *[]byte) [][]int {
	if !r.ignoreCase {
		return r.re.FindAllIndex(b, n)
	}

	if len(*lowerBuf) < len(b) {
		*lowerBuf = make([]byte, len(b)*2)
	}
	transformBuf := (*lowerBuf)[:len(b)]
	BytesToLowerASCII(transformBuf, b)
	return r.re.FindAllIndex(transformBuf, n)
}

func (r *Regexp) Match(b []byte, lowerBuf *[]byte) bool {
	if !r.ignoreCase {
		return r.re.Match(b)
	}

	if len(*lowerBuf) < len(b) {
		*lowerBuf = make([]byte, len(b)*2)
	}
	transformBuf := (*lowerBuf)[:len(b)]
	BytesToLowerASCII(transformBuf, b)
	return r.re.Match(transformBuf)
}
