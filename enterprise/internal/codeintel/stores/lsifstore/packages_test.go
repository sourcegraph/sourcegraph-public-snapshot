package lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/semantic"
)

func TestDatabasePackageInformation(t *testing.T) {
	db, store := testStore(t)
	populateTestIndex(t, db)

	if actual, exists, err := store.PackageInformation(context.Background(), testBundleID, "protocol/protocol.go", "251"); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else if !exists {
		t.Errorf("no package information")
	} else {
		expected := semantic.PackageInformationData{
			Name:    "github.com/sourcegraph/lsif-go",
			Version: "v0.0.0-ad3507cbeb18",
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected package information (-want +got):\n%s", diff)
		}
	}
}
