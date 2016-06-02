package coverageutil

import (
	"testing"
)

func TestJavaScript(testing *testing.T) {
	testTokenizer(testing,
		&javascriptTokenizer{},
		[]*test{
			{
				"backticks and single quotes",
				"`back\ntick`\nconsole.log('hi')",
				[]Token{{12, "console"}, {20, "log"}},
			},
			{
				"double quotes and Unicode code points",
				"a \"\\u{2F804}\" b",
				[]Token{{0, "a"}, {14, "b"}},
			},
			{
				"identifiers",
				"$ = 1; _a = 2;",
				[]Token{{0, "$"}, {7, "_a"}},
			},
			{
				"numeric literals",
				"0b001 0B1 0x0 0X1 000 0o644 0O666 .1 0123 0.04",
				[]Token{},
			},
			{
				"regular expressions and comments",
				"/abc/ /abc/d a / b //abcdef\nccc",
				[]Token{{13, "a"}, {17, "b"}, {28, "ccc"}},
			},
		})
}
