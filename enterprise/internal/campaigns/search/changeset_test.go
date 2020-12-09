package search

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/search/query/syntax"
)

func TestChangesetSearch(t *testing.T) {
	t.Run("parse error", func(t *testing.T) {
		if _, err := ParseChangesetSearch(`:`); err == nil {
			t.Errorf("unexpected nil error")
		}
	})

	t.Run("invalid field", func(t *testing.T) {
		if _, err := ParseChangesetSearch(`x:`); err == nil {
			t.Errorf("unexpected nil error")
		} else if errs, ok := err.(*multierror.Error); !ok {
			t.Errorf("unexpected error of type %T: %+v", err, err)
		} else if diff := cmp.Diff([]error{
			ErrUnsupportedField{
				ErrExpr: ErrExpr{Pos: 0, Input: `x:`},
				Field:   "x",
			},
		}, errs.Errors); diff != "" {
			t.Errorf("unexpected error (-want +have):\n%s", diff)
		}
	})

	t.Run("invalid value type", func(t *testing.T) {
		if _, err := ParseChangesetSearch(`/foo/`); err == nil {
			t.Errorf("unexpected nil error")
		} else if errs, ok := err.(*multierror.Error); !ok {
			t.Errorf("unexpected error of type %T: %+v", err, err)
		} else if diff := cmp.Diff([]error{
			ErrUnsupportedValueType{
				ErrExpr:   ErrExpr{Pos: 1, Input: `/foo/`},
				ValueType: syntax.TokenPattern,
			},
		}, errs.Errors); diff != "" {
			t.Errorf("unexpected error (-want +have):\n%s", diff)
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		if _, err := ParseChangesetSearch(`x: /foo/`); err == nil {
			t.Errorf("unexpected nil error")
		} else if errs, ok := err.(*multierror.Error); !ok {
			t.Errorf("unexpected error of type %T: %+v", err, err)
		} else if diff := cmp.Diff([]error{
			ErrUnsupportedField{
				ErrExpr: ErrExpr{Pos: 0, Input: `x: /foo/`},
				Field:   "x",
			},
			ErrUnsupportedValueType{
				ErrExpr:   ErrExpr{Pos: 4, Input: `x: /foo/`},
				ValueType: syntax.TokenPattern,
			},
		}, errs.Errors); diff != "" {
			t.Errorf("unexpected error (-want +have):\n%s", diff)
		}
	})

	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			input string
			want  campaigns.ListChangesetsOpts
		}{
			"empty string": {
				input: ``,
				want: campaigns.ListChangesetsOpts{
					TextSearch: []campaigns.ListChangesetsTextSearchExpr{},
				},
			},
			"single word": {
				input: `foo`,
				want: campaigns.ListChangesetsOpts{
					TextSearch: []campaigns.ListChangesetsTextSearchExpr{
						{Term: "foo"},
					},
				},
			},
			"negated single word": {
				input: `-foo`,
				want: campaigns.ListChangesetsOpts{
					TextSearch: []campaigns.ListChangesetsTextSearchExpr{
						{Term: "foo", Not: true},
					},
				},
			},
			"quoted phrase": {
				input: `"foo bar"`,
				want: campaigns.ListChangesetsOpts{
					TextSearch: []campaigns.ListChangesetsTextSearchExpr{
						{Term: "foo bar"},
					},
				},
			},
			"negated quoted phrase": {
				input: `-"foo bar"`,
				want: campaigns.ListChangesetsOpts{
					TextSearch: []campaigns.ListChangesetsTextSearchExpr{
						{Term: "foo bar", Not: true},
					},
				},
			},
			"multiple exprs": {
				input: `foo "foo bar" -quux -"baz"`,
				want: campaigns.ListChangesetsOpts{
					TextSearch: []campaigns.ListChangesetsTextSearchExpr{
						{Term: "foo"},
						{Term: "foo bar"},
						{Term: "quux", Not: true},
						{Term: "baz", Not: true},
					},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				opts, err := ParseChangesetSearch(tc.input)
				if err != nil {
					t.Errorf("unexpected error: %+v", err)
				}

				if diff := cmp.Diff(&tc.want, opts); diff != "" {
					t.Errorf("unexpected options (-want +have):\n%s", diff)
				}
			})
		}
	})
}
