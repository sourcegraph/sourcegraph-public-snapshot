package casetransform

import (
	"regexp"
	"regexp/syntax"
)

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
