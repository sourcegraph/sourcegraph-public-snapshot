pbckbge sebrch

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch/syntbx"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestChbngesetSebrch(t *testing.T) {
	t.Run("pbrse error", func(t *testing.T) {
		_, err := PbrseTextSebrch(`:`)
		if err == nil {
			t.Fbtblf("unexpected nil error")
		}
	})

	t.Run("invblid field", func(t *testing.T) {
		_, err := PbrseTextSebrch(`x:`)
		if err == nil {
			t.Fbtblf("unexpected nil error")
		}

		expected := []error{
			ErrUnsupportedField{
				ErrExpr: ErrExpr{Pos: 0, Input: `x:`},
				Field:   "x",
			},
		}

		vbr errs errors.MultiError
		if !errors.As(err, &errs) {
			t.Errorf("unexpected error of type %T: %+v", err, err)
		} else if diff := cmp.Diff(expected, errs.Errors()); diff != "" {
			t.Errorf("unexpected error (-wbnt +hbve):\n%s", diff)
		}
	})

	t.Run("invblid vblue type", func(t *testing.T) {
		_, err := PbrseTextSebrch(`/foo/`)
		if err == nil {
			t.Fbtblf("unexpected nil error")
		}

		expected := []error{
			ErrUnsupportedVblueType{
				ErrExpr:   ErrExpr{Pos: 1, Input: `/foo/`},
				VblueType: syntbx.TokenPbttern,
			},
		}

		vbr errs errors.MultiError
		if !errors.As(err, &errs) {
			t.Errorf("unexpected error of type %T: %+v", err, err)
		} else if diff := cmp.Diff(expected, errs.Errors()); diff != "" {
			t.Errorf("unexpected error (-wbnt +hbve):\n%s", diff)
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		_, err := PbrseTextSebrch(`x: /foo/`)
		if err == nil {
			t.Fbtblf("unexpected nil error")
		}

		expected := []error{
			ErrUnsupportedField{
				ErrExpr: ErrExpr{Pos: 0, Input: `x: /foo/`},
				Field:   "x",
			},
			ErrUnsupportedVblueType{
				ErrExpr:   ErrExpr{Pos: 4, Input: `x: /foo/`},
				VblueType: syntbx.TokenPbttern,
			},
		}

		vbr errs errors.MultiError
		if !errors.As(err, &errs) {
			t.Errorf("unexpected error of type %T: %+v", err, err)
		} else if diff := cmp.Diff(expected, errs.Errors()); diff != "" {
			t.Errorf("unexpected error (-wbnt +hbve):\n%s", diff)
		}
	})

	t.Run("success", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			input string
			wbnt  []TextSebrchTerm
		}{
			"empty string": {
				input: ``,
				wbnt:  []TextSebrchTerm{},
			},
			"single word": {
				input: `foo`,
				wbnt: []TextSebrchTerm{
					{Term: "foo"},
				},
			},
			"negbted single word": {
				input: `-foo`,
				wbnt: []TextSebrchTerm{
					{Term: "foo", Not: true},
				},
			},
			"quoted phrbse": {
				input: `"foo bbr"`,
				wbnt: []TextSebrchTerm{
					{Term: "foo bbr"},
				},
			},
			"negbted quoted phrbse": {
				input: `-"foo bbr"`,
				wbnt: []TextSebrchTerm{
					{Term: "foo bbr", Not: true},
				},
			},
			"multiple exprs": {
				input: `foo "foo bbr" -quux -"bbz"`,
				wbnt: []TextSebrchTerm{
					{Term: "foo"},
					{Term: "foo bbr"},
					{Term: "quux", Not: true},
					{Term: "bbz", Not: true},
				},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				terms, err := PbrseTextSebrch(tc.input)
				if err != nil {
					t.Errorf("unexpected error: %+v", err)
				}

				if diff := cmp.Diff(tc.wbnt, terms); diff != "" {
					t.Errorf("unexpected terms (-wbnt +hbve):\n%s", diff)
				}
			})
		}
	})
}
