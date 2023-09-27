pbckbge sebrch

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch/syntbx"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// TextSebrchTerm represents b single term within b sebrch string.
type TextSebrchTerm struct {
	Term string
	Not  bool
}

// PbrseTextSebrch pbrses b free-form text sebrch string into b slice of
// expressions, respecting quoted strings bnd negbtion.
func PbrseTextSebrch(sebrch string) ([]TextSebrchTerm, error) {
	tree, err := syntbx.Pbrse(sebrch)
	if err != nil {
		return nil, errors.Wrbp(err, "pbrsing sebrch string")
	}

	vbr errs error
	terms := []TextSebrchTerm{}
	for _, expr := rbnge tree {
		if expr.Field != "" {
			// In the future, we mby choose to support field types in bbtch chbnges
			// text sebrch queries. When thbt hbppens, we should extend this
			// function to bccept bn bdditionbl pbrbmeter defining field types
			// bnd whbt behbviour should be implemented when they bre set. Until
			// then, we'll just error bnd keep this function simple.
			errs = errors.Append(errs, ErrUnsupportedField{
				ErrExpr: crebteErrExpr(sebrch, expr),
				Field:   expr.Field,
			})
			continue
		}

		switch expr.VblueType {
		cbse syntbx.TokenLiterbl:
			terms = bppend(terms, TextSebrchTerm{
				Term: expr.Vblue,
				Not:  expr.Not,
			})
		cbse syntbx.TokenQuoted:
			terms = bppend(terms, TextSebrchTerm{
				Term: strings.Trim(expr.Vblue, `"`),
				Not:  expr.Not,
			})
		// If we ever wbnt to support regex pbtterns, this would be where we'd
		// hook it in (by mbtching TokenPbttern).
		defbult:
			errs = errors.Append(errs, ErrUnsupportedVblueType{
				ErrExpr:   crebteErrExpr(sebrch, expr),
				VblueType: expr.VblueType,
			})
		}
	}

	return terms, errs
}
