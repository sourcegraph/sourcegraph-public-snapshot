package coverageutil

import (
	"testing"
)

func TestCSS(testing *testing.T) {

	type test struct {
		name     string
		source   string
		expected []Token
	}
	tests := []*test{
		{
			"UTF-8",
			"td { content: \"©\"; color: red }\n tr {font-size:11px;}",
			[]Token{{0, 1, "td"}, {5, 1, "content"}, {20, 1, "color"}, {34, 2, "tr"}, {38, 2, "font-size"}},
		},
		// TODO(alexsaveliev) remove/rewrite when sourcegraph/srclib-css#6 will be fixed.
		// Please note that while it's expected behavior to generate {0, "c"} token
		// {1, "b} is still wrong and should be fixed in sourcegraph/srclib-css#6
		// After the fix, expected output is {0, "a"}, {3, "b"}, and {6, "c"}
		{
			"Wrong offsets",
			"a,\nb,\nc {}",
			[]Token{{0, 1, "a"}, {1, 1, "b"}, {0, 1, "c"}},
		},
		// TODO(alexsaveliev) uncomment when parser will not include comments into decls and selectors
		// See sourcegraph/srclib-css#5
		//
		//		{
		//			"Comments",
		//			"td /* © */ { color /* © */: red; }",
		//			[]Token{{0, "td"}, {14, "color"}},
		//		},
	}
	tokenizer := &cssTokenizer{}
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
			if tok.Offset != t.expected[i].Offset || tok.Line != t.expected[i].Line || tok.Text != t.expected[i].Text {
				testing.Errorf("%s: Expected %d %d (%s), got %d %d (%s) instead", t.name, t.expected[i].Offset, t.expected[i].Line, t.expected[i].Text, tok.Offset, tok.Line, tok.Text)
			}
		}
	}
}
