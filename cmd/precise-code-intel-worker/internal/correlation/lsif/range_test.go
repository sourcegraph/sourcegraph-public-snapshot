package lsif

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"
)

func TestUnmarshalRangeData(t *testing.T) {
	r, err := UnmarshalRangeData(Element{
		ID:    "04",
		Type:  "vertex",
		Label: "range",
		Raw:   json.RawMessage(`{"id": "04", "type": "vertex", "label": "range", "start": {"line": 1, "character": 2}, "end": {"line": 3, "character": 4}}`),
	})
	if err != nil {
		t.Fatalf("unexpected error unmarshalling range data: %s", err)
	}

	expectedRange := RangeData{
		StartLine:          1,
		StartCharacter:     2,
		EndLine:            3,
		EndCharacter:       4,
		DefinitionResultID: "",
		ReferenceResultID:  "",
		HoverResultID:      "",
		MonikerIDs:         datastructures.IDSet{},
	}
	if diff := cmp.Diff(expectedRange, r); diff != "" {
		t.Errorf("unexpected range (-want +got):\n%s", diff)
	}
}
