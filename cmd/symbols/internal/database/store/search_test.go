package store

import "testing"

func TestIsLiteralEquality(t *testing.T) {
	for _, test := range []struct {
		regex           string
		noMatch         bool
		expectedLiteral string
	}{
		{regex: `^foo$`, expectedLiteral: "foo"},
		{regex: `^[f]oo$`, expectedLiteral: `foo`},
		{regex: `^\\$`, expectedLiteral: `\`},
		{regex: `^\$`, noMatch: true},
		{regex: `^\($`, expectedLiteral: `(`},
		{regex: `\\`, noMatch: true},
		{regex: `\$`, noMatch: true},
		{regex: `\(`, noMatch: true},
		{regex: `foo$`, noMatch: true},
		{regex: `(^foo$|^bar$)`, noMatch: true},
	} {
		literal, ok, err := isLiteralEquality(test.regex)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			if !test.noMatch {
				t.Errorf("exected a match")
			}
		} else if test.noMatch {
			t.Errorf("did not expect a match")
		} else if literal != test.expectedLiteral {
			t.Errorf(
				"unexpected literal for %q. want=%q have=%q",
				test.regex,
				test.expectedLiteral,
				literal,
			)
		}
	}
}

func TestIsLiteralPrefix(t *testing.T) {
	for _, test := range []struct {
		regex           string
		noMatch         bool
		expectedLiteral string
	}{
		{regex: `^foo`, expectedLiteral: "foo"},
		{regex: `^[f]oo`, expectedLiteral: `foo`},
		{regex: `^\\`, expectedLiteral: `\`},
		{regex: `^\(`, expectedLiteral: `(`},
		{regex: `\\`, noMatch: true},
		{regex: `\$`, noMatch: true},
		{regex: `\(`, noMatch: true},
		{regex: `foo$`, noMatch: true},
		{regex: `(^foo$|^bar$)`, noMatch: true},
	} {
		literal, ok, err := isLiteralPrefix(test.regex)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			if !test.noMatch {
				t.Errorf("exected a match")
			}
		} else if test.noMatch {
			t.Errorf("did not expect a match")
		} else if literal != test.expectedLiteral {
			t.Errorf(
				"unexpected literal for %q. want=%q have=%q",
				test.regex,
				test.expectedLiteral,
				literal,
			)
		}
	}
}

func TestEscapeGlob(t *testing.T) {
	for _, test := range []struct {
		str  string
		want string
	}{
		{str: "", want: ""},
		{str: "foo", want: "foo"},
		{str: "*", want: "[*]"},
		{str: "?", want: "[?]"},
		{str: "[", want: "[[]"},
		{str: "]", want: "[]]"},
		{str: "**?foo]*[", want: "[*][*][?]foo[]][*][[]"},
	} {
		got := globEscape(test.str)
		if got != test.want {
			t.Errorf(
				"unexpected result for escaping %q. want=%q got=%q",
				test.str,
				test.want,
				got,
			)
		}
	}
}
