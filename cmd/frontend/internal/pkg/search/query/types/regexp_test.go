package types

import (
	"regexp/syntax"
	"testing"
)

func TestFixupCompileErrors(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{query: "$foo[", want: `\$foo\[`},
		{query: "foo(", want: `foo\(`},
		{query: "foo[", want: `foo\[`},
		{query: "*foo", want: `\*foo`},
		{query: "$foo", want: `\$foo`},
		{query: `foo\s=\s$bar`, want: `foo\s=\s\$bar`},

		// Valid regexps
		{query: `foo\(`, want: `foo\(`},
		{query: `foo\[`, want: `foo\[`},
		{query: `\*foo`, want: `\*foo`},
		{query: `\$foo`, want: `\$foo`},
		{query: `foo$`, want: `foo$`},
		{query: `foo\s=\s\$bar`, want: `foo\s=\s\$bar`},
		{query: "[$]", want: `[$]`},
	}

	for _, test := range tests {
		got, err := parseRegexp(test.query)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		want, _ := syntax.Parse(test.want, syntax.Perl)
		if got.String() != want.String() {
			t.Errorf("query %s got %s want %s", test.query, got, want)
		}
	}
}

func TestFixupCompileErrors_failures(t *testing.T) {
	tests := []string{
		// If the user is trying to use capture groups, then forgetting to escape a paren is definitely an error.
		"(foo|bar)(",
	}

	for _, query := range tests {
		_, gotErr := parseRegexp(query)
		if gotErr == nil {
			t.Errorf("expected error for `%s`, got none", query)
			continue
		}

		_, wantErr := syntax.Parse(query, syntax.Perl)
		if gotErr.Error() != wantErr.Error() {
			t.Errorf("error for query %s was different than expected\ngot: %v\nwanted: %v", query, gotErr, wantErr)
		}
	}
}
