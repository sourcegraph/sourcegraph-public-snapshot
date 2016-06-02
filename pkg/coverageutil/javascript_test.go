package coverageutil

import (
	"bytes"
	"testing"
	"text/scanner"
)

func TestJavaScript(testing *testing.T) {

	type test struct {
		name     string
		source   string
		expected []Token
	}
	tests := []*test{
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
	}
	jsScanner := newJavascriptScanner()
	for _, t := range tests {
		jsScanner.Init(bytes.NewReader([]byte(t.source)))
		actual := make([]Token, 0)
		for {
			tok := jsScanner.Scan()
			if tok == scanner.EOF {
				break
			}
			text := jsScanner.TokenText()
			actual = append(actual, Token{uint32(jsScanner.Pos().Offset - len([]byte(text))), text})
		}
		if len(actual) != len(t.expected) {
			testing.Fatalf("%s: Expected %d tokens, got %d instead", t.name, len(t.expected), len(actual))
		}
		for i, tok := range actual {
			if tok.Offset != t.expected[i].Offset || tok.Text != t.expected[i].Text {
				testing.Errorf("%s: Expected %d (%s), got %d (%s) instead", t.name, t.expected[i].Offset, t.expected[i].Text, tok.Offset, tok.Text)
			}
		}
	}
}
