package fix

import (
	"testing"
)

func TestFixableRegexp(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{"$foo", `\$foo`},
		{"$foo[", `\$foo\[`},
		{"foo(", `foo\(`},
		{"foo[", `foo\[`},
		{"*foo", `\*foo`},
		{"$foo", `\$foo`},
		{`foo\s=\s$bar`, `foo[\t-\n\f-\r ]=[\t-\n\f-\r ]\$bar`},
		{"foo)", `foo\)`},
		{"foo]", `foo\]`},

		// Valid regexps
		{"^f.*o$", "^f.*o$"},
		{"$foo", `\$foo`},
		{`foo\(`, `foo\(`},
		{`foo\[`, `foo\[`},
		{`\*foo`, `\*foo`},
		{`\$foo`, `\$foo`},
		{`foo$`, `foo$`},
		{`foo\s=\s\$bar`, `foo[\t-\n\f-\r ]=[\t-\n\f-\r ]\$bar`},
		{"[$]", `\$`}, // Intentionally don't support only matching end of line/text because it is not useful in a code search
	}

	for _, test := range tests {
		fr := NewFixableRegexp(test.input)

		fr.Fix()

		if got := fr.String(); got != test.output {
			t.Errorf("input tranformed in an unexpected way\ngot: %v\nwant: %v\n", got, test.output)
		}
	}
}
