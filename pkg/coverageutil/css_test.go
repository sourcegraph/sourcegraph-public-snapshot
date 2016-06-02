package coverageutil

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
				[]Token{{0, 1, "td"}, {5, 1, "content"}, {20, 1, "color"}, {34, 2, "tr"}, {38, 2, "font-size"}},
			},
			{
				"Comma-separated declarations",
				"a,\nb,\nc {}",
				[]Token{{0, 1, "a"}, {3, 2, "b"}, {6, 3, "c"}},
			},
			{
				"Comments",
				"td /* © */ { color /* © */: red; }",
				[]Token{{0, 1, "td"}, {14, 1, "color"}},
			},
		})
}
