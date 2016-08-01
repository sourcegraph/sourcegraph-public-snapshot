package tokenizer

import (
	"testing"
)

func TestCSS(testing *testing.T) {
	testTokenizer(testing,
		&cssTokenizer{},
		[]*test{
			{
				"UTF-8",
				"td { content: \"©\"; color: red }\n tr {font-size:11px;}",
				[]Token{{0, 1, 1, "td"}, {5, 1, 6, "content"}, {20, 1, 20, "color"}, {34, 2, 2, "tr"}, {38, 2, 6, "font-size"}},
			},
			{
				"Comma-separated declarations",
				"a,\nb,\nc {}",
				[]Token{{0, 1, 1, "a"}, {3, 2, 1, "b"}, {6, 3, 1, "c"}},
			},
			{
				"Comments",
				"td /* © */ { color /* © */: red; }",
				[]Token{{0, 1, 1, "td"}, {14, 1, 14, "color"}},
			},
		})
}
