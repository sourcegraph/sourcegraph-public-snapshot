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
			[]Token{{12, 3, "console"}, {20, 3, "log"}},
		},
		{
			"double quotes and Unicode code points",
			"a \"\\u{2F804}\" b",
			[]Token{{0, 1, "a"}, {14, 1, "b"}},
		},
		{
			"identifiers",
			"$ = 1; _a = 2;",
			[]Token{{0, 1, "$"}, {7, 1, "_a"}},
		},
		{
			"numeric literals",
			"0b001 0B1 0x0 0X1 000 0o644 0O666 .1 0123 0.04",
			[]Token{},
		},
		{
			"regular expressions",
			"/abc/ /abc/d a / b",
			[]Token{{13, 1, "a"}, {17, 1, "b"}},
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
			actual = append(actual, Token{uint32(jsScanner.Pos().Offset - len([]byte(text))), jsScanner.Line, text})
		}
		if len(actual) != len(t.expected) {
			testing.Fatalf("%s: Expected %d tokens, got %d instead", t.name, len(t.expected), len(actual))
		}
		for i, tok := range actual {
			if tok.Offset != t.expected[i].Offset || tok.Line != t.expected[i].Line || tok.Text != t.expected[i].Text {
				testing.Errorf("%s: Expected %d %d (%s), got %d %d (%s) instead", t.name, t.expected[i].Offset, t.expected[i].Line, t.expected[i].Text, tok.Offset, tok.Line, tok.Text)
			}
		}
	}
}
