package lsif

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalElement(t *testing.T) {
	raw := `{"id": "47", "type": "edge", "label": "contains", "outV": "02", "inVs": ["04", "05", "06"]}`

	element, err := UnmarshalElement([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error unmarshalling element data: %s", err)
	}

	expectedElement := Element{
		ID:    "47",
		Type:  "edge",
		Label: "contains",
		Raw:   json.RawMessage(raw),
	}
	if diff := cmp.Diff(expectedElement, element); diff != "" {
		t.Errorf("unexpected element (-want +got):\n%s", diff)
	}
}

func TestUnmarshalEdge(t *testing.T) {
	edge, err := UnmarshalEdge(Element{
		ID:    "35",
		Type:  "edge",
		Label: "item",
		Raw:   json.RawMessage(`{"id": "35", "type": "edge", "label": "item", "outV": "12", "inVs": ["07"], "document": "03"}`),
	})
	if err != nil {
		t.Fatalf("unexpected error unmarshalling meta data: %s", err)
	}

	expectedEdge := Edge{
		OutV:     "12",
		InV:      "",
		InVs:     []string{"07"},
		Document: "03",
	}
	if diff := cmp.Diff(expectedEdge, edge); diff != "" {
		t.Errorf("unexpected edge (-want +got):\n%s", diff)
	}
}
