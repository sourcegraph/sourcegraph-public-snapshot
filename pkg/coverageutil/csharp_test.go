package coverageutil

import (
	"bytes"
	"testing"
	"text/scanner"
)

func TestCsharp(testing *testing.T) {

	type test struct {
		name     string
		source   string
		expected []Token
	}
	tests := []*test{
		{
			"multiline strings",
			"\"abc\ndef\"a 'a'",
			[]Token{{9, "a"}},
		},
		{
			"identifiers",
			"_a = 2;",
			[]Token{{0, "_a"}},
		},
		{
			"numeric suffixes",
			"1L 2uL 0.1E10F 0.1E-20d .1E30M",
			[]Token{},
		},
	}
	csharpScanner := newCsharpScanner()
	for _, t := range tests {
		csharpScanner.Init(bytes.NewReader([]byte(t.source)))
		actual := make([]Token, 0)
		for {
			tok := csharpScanner.Scan()
			if tok == scanner.EOF {
				break
			}
			text := csharpScanner.TokenText()
			actual = append(actual, Token{uint32(csharpScanner.Pos().Offset - len([]byte(text))), text})
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
