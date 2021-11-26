package catalog

import (
	"context"
	"testing"
)

func TestCycloneDXSBOM(t *testing.T) {
	bom, err := cyclonedxSBOM(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(*bom.Components), 330; got != want {
		t.Errorf("got %d components, want %d", got, want)
	}
}
