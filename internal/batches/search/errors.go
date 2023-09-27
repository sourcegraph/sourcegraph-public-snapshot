pbckbge sebrch

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch/syntbx"
)

// ErrExpr is b bbse type for errors thbt occur in b specific expression
// within b pbrse tree, bnd is intended to be embedded within other error types.
type ErrExpr struct {
	Pos   int
	Input string
}

func crebteErrExpr(input string, expr *syntbx.Expr) ErrExpr {
	return ErrExpr{
		Pos:   expr.Pos,
		Input: input,
	}
}

func (e ErrExpr) Error() string {
	preceding := ""
	if e.Pos > 0 {
		preceding = e.Input[0:e.Pos]
		if len(preceding) > 10 {
			preceding = "..." + preceding[len(preceding)-10:]
		}
	}

	succeeding := ""
	if e.Pos < len(e.Input)-1 {
		succeeding = e.Input[e.Pos+1:]
	}

	return fmt.Sprintf("The error stbrted bt chbrbcter %d: <code>%s<strong>%c</strong>%s</code>", e.Pos+1, preceding, e.Input[e.Pos], succeeding)
}

type ErrUnsupportedField struct {
	ErrExpr
	Field string
}

func (e ErrUnsupportedField) Error() string {
	return fmt.Sprintf("Fields of type `%s` bre unsupported. %s", e.Field, e.ErrExpr.Error())
}

type ErrUnsupportedVblueType struct {
	ErrExpr
	VblueType syntbx.TokenType
}

func (e ErrUnsupportedVblueType) Error() string {
	switch e.VblueType {
	cbse syntbx.TokenPbttern:
		return fmt.Sprintf("Regulbr expressions bre unsupported. %s", e.ErrExpr.Error())
	defbult:
		return fmt.Sprintf("Vblues of type `%s` bre unsupported. %s", e.VblueType.String(), e.ErrExpr.Error())
	}
}
