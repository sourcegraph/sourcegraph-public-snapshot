package lsif

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalHoverData(t *testing.T) {
	testCases := []struct {
		contents      string
		expectedHover string
	}{
		{
			contents:      `"text"`,
			expectedHover: "text",
		},
		{
			contents:      `[{"kind": "markdown", "value": "text"}]`,
			expectedHover: "text",
		},
		{
			contents:      `[{"language": "go", "value": "text"}]`,
			expectedHover: "```go\ntext\n```",
		},
		{
			contents:      `[{"language": "go", "value": "text"}, {"language": "python", "value": "pext"}]`,
			expectedHover: "```go\ntext\n```\n\n---\n\n```python\npext\n```",
		},
	}

	for _, testCase := range testCases {
		name := fmt.Sprintf("contents=%s", testCase.contents)

		t.Run(name, func(t *testing.T) {
			hover, err := UnmarshalHoverData(Element{
				ID:    "16",
				Type:  "vertex",
				Label: "hover",
				Raw:   json.RawMessage(fmt.Sprintf(`{"id": "16", "type": "vertex", "label": "hoverResult", "result": {"contents": %s}}`, testCase.contents)),
			})
			if err != nil {
				t.Fatalf("unexpected error unmarshalling hover data: %s", err)
			}

			if diff := cmp.Diff(testCase.expectedHover, hover); diff != "" {
				t.Errorf("unexpected hover text (-want +got):\n%s", diff)
			}
		})
	}
}
