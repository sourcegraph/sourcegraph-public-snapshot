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
				[]Token{{0, "td"}, {5, "content"}, {20, "color"}, {34, "tr"}, {38, "font-size"}},
			},
			{
				"Comma-separated declarations",
				"a,\nb,\nc {}",
				[]Token{{0, "a"}, {3, "b"}, {6, "c"}},
			},
			{
				"Comments",
				"td /* © */ { color /* © */: red; }",
				[]Token{{0, "td"}, {14, "color"}},
			},
		})
}
