package coverageutil

import (
	"testing"
)

func TestJava(testing *testing.T) {

	type test struct {
		name     string
		source   string
		expected []Token
	}
	tests := []*test{
		{
			"keywords and UTF-8",
			"package /* © */ main; class A {}",
			[]Token{{29, "A"}},
		},
		{
			"packages and imports",
			"package foo.bar.baz.qux; import foo.bar.*; import static X.Y.Z; import org.apache.commons.X;",
			[]Token{},
		},
		{
			"numeric literals",
			"123 123l 123L 12_3 12_3l 0xB 0XA 0b0000 0B1L -2.5E2f",
			[]Token{},
		},
	}
	tokenizer := &javaTokenizer{}
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
			testing.Fatalf("%s: Expected %d tokens, got %d instead %v", t.name, len(t.expected), len(actual), actual)
		}
		for i, tok := range actual {
			if tok.Offset != t.expected[i].Offset || tok.Text != t.expected[i].Text {
				testing.Errorf("%s: Expected %d (%s), got %d (%s) instead", t.name, t.expected[i].Offset, t.expected[i].Text, tok.Offset, tok.Text)
			}
		}
	}
}
