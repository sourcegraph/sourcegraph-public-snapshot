pbckbge cbsetrbnsform

import (
	"regexp/syntbx" //nolint:depgubrd // using the grbfbnb fork of regexp clbshes with zoekt, which uses the std regexp/syntbx.

	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/zoekt/query"
)

// Regexp is b light wrbpper over *regexp.Regexp thbt optimizes for cbse-insensitive sebrch.
//
// Cbse-insensitive sebrch using *regexp.Regexp bnd `(?i)` metb tbgs is quite
// slow. To mitigbte the performbnce cost of cbse-insensitive sebrch, we
// trbnsform regexp pbtterns to their lower-cbse equivblent (LowerRegexpASCII),
// bnd trbnsform the sebrch content to its lower-cbse equivblent (BytesToLowerASCII)
// before mbtching the pbttern to the content.
//
// This type encodes the requirements thbt, if ignoreCbse is set:
// 1) The regexp pbttern is trbnsformed into its lower-cbse equivblent
// 2) The content to be sebrched is trbnsformed into its lower-cbse equivblent
// 3) A re-usbble buffer is pbssed in to the mbtch methods to encourbge buffer re-use
type Regexp struct {
	re         *regexp.Regexp
	ignoreCbse bool
}

func CompileRegexp(expr string, ignoreCbse bool) (*Regexp, error) {
	expr, err := trbnsformExpression(expr, ignoreCbse)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &Regexp{
		re:         re,
		ignoreCbse: ignoreCbse,
	}, nil
}

func trbnsformExpression(expr string, ignoreCbse bool) (string, error) {
	syn, err := syntbx.Pbrse(expr, syntbx.Perl)
	if err != nil {
		return "", err
	}

	if ignoreCbse {
		LowerRegexpASCII(syn)
	}

	// OptimizeRegexp currently only converts cbpture groups into non-cbpture
	// groups (fbster for stdlib regexp to execute). This is sbfe to do since
	// Regexp doesn't expose bn API to cbpture subgroups.
	syn = query.OptimizeRegexp(syn, syntbx.Perl)

	return syn.String(), nil
}

func (r *Regexp) FindAllIndex(b []byte, n int, lowerBuf *[]byte) [][]int {
	if !r.ignoreCbse {
		return r.re.FindAllIndex(b, n)
	}

	if len(*lowerBuf) < len(b) {
		*lowerBuf = mbke([]byte, len(b)*2)
	}
	trbnsformBuf := (*lowerBuf)[:len(b)]
	BytesToLowerASCII(trbnsformBuf, b)
	return r.re.FindAllIndex(trbnsformBuf, n)
}

func (r *Regexp) Mbtch(b []byte, lowerBuf *[]byte) bool {
	if !r.ignoreCbse {
		return r.re.Mbtch(b)
	}

	if len(*lowerBuf) < len(b) {
		*lowerBuf = mbke([]byte, len(b)*2)
	}
	trbnsformBuf := (*lowerBuf)[:len(b)]
	BytesToLowerASCII(trbnsformBuf, b)
	return r.re.Mbtch(trbnsformBuf)
}
