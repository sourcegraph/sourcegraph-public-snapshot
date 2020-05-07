package lsif

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalMetaData(t *testing.T) {
	element := Element{
		ID:    "01",
		Type:  "vertex",
		Label: "metaData",
		Raw:   json.RawMessage(`{"id": "01", "type": "vertex", "label": "metaData", "version": "0.4.3", "projectRoot": "file:///test"}`),
	}

	metadata, err := UnmarshalMetaData(element, "foo/bar")
	if err != nil {
		t.Fatalf("unexpected error unmarshalling meta data: %s", err)
	}

	expectedMetadata := MetaData{
		Version:     "0.4.3",
		ProjectRoot: "file:///test/foo/bar",
	}
	if diff := cmp.Diff(expectedMetadata, metadata); diff != "" {
		t.Errorf("unexpected metadata (-want +got):\n%s", diff)
	}
}
