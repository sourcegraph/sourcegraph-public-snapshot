package coverageutil

import (
	"bytes"
	"testing"
	"text/scanner"
)

func TestPython(testing *testing.T) {

	type test struct {
		name     string
		source   string
		expected []Token
	}
	tests := []*test{
		{
			"strings",
			"'abc' \"abc\" '''abc''' \"\"\"abc\"\"\" '''abc",
			[]Token{},
		},
		{
			"raw and Unicode strings",
			"r'abc' rU\"abc\" U'''abc''' uR\"\"\"abc\"\"\" UR'''abc",
			[]Token{},
		},
		{
			"complex numbers",
			"2+10j 2+10q",
			[]Token{{10, "q"}},
		},
		{
			"comments",
			"#abc\ndef#fgh",
			[]Token{{5, "def"}},
		},
	}
	pythonScanner := newPythonScanner()
	for _, t := range tests {
		pythonScanner.Init(bytes.NewReader([]byte(t.source)))
		actual := make([]Token, 0)
		for {
			tok := pythonScanner.Scan()
			if tok == scanner.EOF {
				break
			}
			text := pythonScanner.TokenText()
			actual = append(actual, Token{uint32(pythonScanner.Pos().Offset - len([]byte(text))), text})
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
