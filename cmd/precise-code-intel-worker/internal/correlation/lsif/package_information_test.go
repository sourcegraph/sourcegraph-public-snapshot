package lsif

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalPackageInformationData(t *testing.T) {
	packageInformation, err := UnmarshalPackageInformationData(Element{
		ID:    "22",
		Type:  "vertex",
		Label: "packageInformation",
		Raw:   json.RawMessage(`{"id": "22", "type": "vertex", "label": "packageInformation", "name": "pkg A", "version": "v0.1.0"}`),
	})
	if err != nil {
		t.Fatalf("unexpected error unmarshalling package information data: %s", err)
	}

	expectedPackageInformation := PackageInformationData{
		Name:    "pkg A",
		Version: "v0.1.0",
	}
	if diff := cmp.Diff(expectedPackageInformation, packageInformation); diff != "" {
		t.Errorf("unexpected package information (-want +got):\n%s", diff)
	}
}
