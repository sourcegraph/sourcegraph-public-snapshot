package lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func TestDatabasePackageInformation(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	populateTestStore(t)
	store := NewStore(db, &observation.TestContext)

	if actual, exists, err := store.PackageInformation(context.Background(), testBundleID, "protocol/protocol.go", "251"); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else if !exists {
		t.Errorf("no package information")
	} else {
		expected := precise.PackageInformationData{
			Name:    "github.com/sourcegraph/lsif-go",
			Version: "v0.0.0-ad3507cbeb18",
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected package information (-want +got):\n%s", diff)
		}
	}
}
