package lsif

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalMonikerData(t *testing.T) {
	moniker, err := UnmarshalMonikerData(Element{
		ID:    "18",
		Type:  "vertex",
		Label: "moniker",
		Raw:   json.RawMessage(`{"id": "18", "type": "vertex", "label": "moniker", "kind": "import", "scheme": "scheme A", "identifier": "ident A"}`),
	})
	if err != nil {
		t.Fatalf("unexpected error unmarshalling moniker data: %s", err)
	}

	expectedMoniker := MonikerData{
		Kind:       "import",
		Scheme:     "scheme A",
		Identifier: "ident A",
	}
	if diff := cmp.Diff(expectedMoniker, moniker); diff != "" {
		t.Errorf("unexpected moniker (-want +got):\n%s", diff)
	}
}
