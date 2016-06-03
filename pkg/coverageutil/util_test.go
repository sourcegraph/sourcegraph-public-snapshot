package coverageutil

import (
	"testing"
)

type test struct {
	name     string
	source   string
	expected []Token
}

// testTokenizer runs tests using the given tokenizer
// to ensure it produces expected results
func testTokenizer(testing *testing.T, tokenizer Tokenizer, tests []*test) {
	for _, t := range tests {
		tokenizer.Init([]byte(t.source))
		defer tokenizer.Done()
		actual := make([]*Token, 0)
		for {
			tok := tokenizer.Next()
			if tok == nil {
				break
			}
			actual = append(actual, tok)
		}
		errors := tokenizer.Errors()
		if len(errors) != 0 {
			testing.Fatalf("%s: Got errors %v", t.name, errors)
		}
		if len(actual) != len(t.expected) {
			testing.Fatalf("%s: Expected %d tokens, got %d instead", t.name, len(t.expected), len(actual))
		}
		for i, tok := range actual {
			if tok.Offset != t.expected[i].Offset || tok.Line != t.expected[i].Line || tok.Text != t.expected[i].Text {
				testing.Errorf("%s: Expected %d,%d (%s), got %d,%d (%s) instead", t.name, t.expected[i].Offset, t.expected[i].Line, t.expected[i].Text, tok.Offset, tok.Line, tok.Text)
			}
		}
	}
}
