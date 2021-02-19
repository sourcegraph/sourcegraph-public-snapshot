package dbstore

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

func TestUpdatePackages(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	// for foreign key relation
	insertUploads(t, dbconn.Global, Upload{ID: 42})

	if err := store.UpdatePackages(context.Background(), []lsifstore.Package{
		{DumpID: 42, Scheme: "s0", Name: "n0", Version: "v0"},
		{DumpID: 42, Scheme: "s1", Name: "n1", Version: "v1"},
		{DumpID: 42, Scheme: "s2", Name: "n2", Version: "v2"},
		{DumpID: 42, Scheme: "s3", Name: "n3", Version: "v3"},
		{DumpID: 42, Scheme: "s4", Name: "n4", Version: "v4"},
		{DumpID: 42, Scheme: "s5", Name: "n5", Version: "v5"},
		{DumpID: 42, Scheme: "s6", Name: "n6", Version: "v6"},
		{DumpID: 42, Scheme: "s7", Name: "n7", Version: "v7"},
		{DumpID: 42, Scheme: "s8", Name: "n8", Version: "v8"},
		{DumpID: 42, Scheme: "s9", Name: "n9", Version: "v9"},
	}); err != nil {
		t.Fatalf("unexpected error updating packages: %s", err)
	}

	count, _, err := basestore.ScanFirstInt(dbconn.Global.Query("SELECT COUNT(*) FROM lsif_packages"))
	if err != nil {
		t.Fatalf("unexpected error checking package count: %s", err)
	}
	if count != 10 {
		t.Errorf("unexpected package count. want=%d have=%d", 10, count)
	}
}

func TestUpdatePackagesEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	if err := store.UpdatePackages(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error updating packages: %s", err)
	}

	count, _, err := basestore.ScanFirstInt(dbconn.Global.Query("SELECT COUNT(*) FROM lsif_packages"))
	if err != nil {
		t.Fatalf("unexpected error checking package count: %s", err)
	}
	if count != 0 {
		t.Errorf("unexpected package count. want=%d have=%d", 0, count)
	}
}
