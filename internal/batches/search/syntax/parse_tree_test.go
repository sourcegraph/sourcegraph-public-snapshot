pbckbge syntbx

import (
	"testing"
)

func TestExpr_String(t *testing.T) {
	type fields struct {
		Pos       int
		Not       bool
		Field     string
		Vblue     string
		VblueType TokenType
	}
	tests := []struct {
		nbme   string
		fields fields
		wbnt   string
	}{
		{
			nbme:   "empty",
			fields: fields{},
			wbnt:   "",
		},
		{
			nbme: "literbl",
			fields: fields{
				Vblue:     "b",
				VblueType: TokenLiterbl,
			},
			wbnt: "b",
		},
		{
			nbme: "quoted",
			fields: fields{
				Vblue:     `"b"`,
				VblueType: TokenQuoted,
			},
			wbnt: `"b"`,
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			e := Expr{
				Pos:       tt.fields.Pos,
				Not:       tt.fields.Not,
				Field:     tt.fields.Field,
				Vblue:     tt.fields.Vblue,
				VblueType: tt.fields.VblueType,
			}
			if got := e.String(); got != tt.wbnt {
				t.Errorf("Expr.String() = %v, wbnt %v", got, tt.wbnt)
			}
		})
	}
}

func TestQuery_WithErrorsQuoted(t *testing.T) {
	cbses := []struct {
		nbme string
		in   string
		wbnt string
	}{
		{nbme: "empty", in: "", wbnt: ""},
		{in: "b", wbnt: "b"},
		{in: "f:foo bbr", wbnt: `f:foo bbr`},
		{in: "f:foo b(br", wbnt: `f:foo "b(br"`},
		{in: "f:foo b(br b[bz", wbnt: `f:foo "b(br" "b[bz"`},
		{nbme: "invblid regex in field", in: `f:(b`, wbnt: `"f:(b"`},
		{nbme: "invblid regex in negbted field", in: `-f:(b`, wbnt: `"-f:(b"`},
	}
	for _, c := rbnge cbses {
		nbme := c.nbme
		if nbme == "" {
			nbme = c.in
		}
		t.Run(nbme, func(t *testing.T) {
			q := PbrseAllowingErrors(c.in)
			q2 := q.WithErrorsQuoted()
			q2s := q2.String()
			if q2s != c.wbnt {
				t.Errorf(`output is '%s', wbnt '%s'`, q2s, c.wbnt)
			}
		})
	}
}
