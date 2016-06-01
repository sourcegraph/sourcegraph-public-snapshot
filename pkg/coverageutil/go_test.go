package coverageutil

import (
	"testing"
)

func TestGo(testing *testing.T) {

	type test struct {
		name     string
		source   string
		expected []Token
	}
	tests := []*test{
		{
			"keywords",
			"package main",
			[]Token{{8, "main"}},
		},
	}
	tokenizer := &goTokenizer{}
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
