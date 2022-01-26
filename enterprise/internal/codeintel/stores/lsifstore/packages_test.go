package lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func TestDatabasePackageInformation(t *testing.T) {
	store := populateTestStore(t)

	if actual, exists, err := store.PackageInformation(context.Background(), testBundleID, "protocol/protocol.go", "114"); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else if !exists {
		t.Errorf("no package information")
	} else {
		expected := precise.PackageInformationData{
			Name:    "https://github.com/sourcegraph/lsif-go",
			Version: "ad3507cbeb18",
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected package information (-want +got):\n%s", diff)
		}
	}
}
