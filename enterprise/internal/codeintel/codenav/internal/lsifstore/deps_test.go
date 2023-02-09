package lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetDependencies(t *testing.T) {
	store := populateTestStore(t)

	dependencies, err := store.GetDependencies(context.Background(), testSCIPUploadID)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	expectedDependencies := []DependencyDescription{
		// TODO
	}
	if diff := cmp.Diff(dependencies, expectedDependencies); diff != "" {
		t.Errorf("unexpected dependencies (-want +got):\n%s", diff)
	}
}
