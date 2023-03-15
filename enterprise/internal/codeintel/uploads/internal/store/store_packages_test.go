package store

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func TestUpdatePackages(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	// for foreign key relation
	insertUploads(t, db, types.Upload{ID: 42})

	if err := store.UpdatePackages(context.Background(), 42, []precise.Package{
		{Scheme: "s0", Name: "n0", Version: "v0"},
		{Scheme: "s1", Name: "n1", Version: "v1"},
		{Scheme: "s2", Name: "n2", Version: "v2"},
		{Scheme: "s3", Name: "n3", Version: "v3"},
		{Scheme: "s4", Name: "n4", Version: "v4"},
		{Scheme: "s5", Name: "n5", Version: "v5"},
		{Scheme: "s6", Name: "n6", Version: "v6"},
		{Scheme: "s7", Name: "n7", Version: "v7"},
		{Scheme: "s8", Name: "n8", Version: "v8"},
		{Scheme: "s9", Name: "n9", Version: "v9"},
	}); err != nil {
		t.Fatalf("unexpected error updating packages: %s", err)
	}

	count, _, err := basestore.ScanFirstInt(db.QueryContext(context.Background(), "SELECT COUNT(*) FROM lsif_packages"))
	if err != nil {
		t.Fatalf("unexpected error checking package count: %s", err)
	}
	if count != 10 {
		t.Errorf("unexpected package count. want=%d have=%d", 10, count)
	}
}

func TestUpdatePackagesEmpty(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	if err := store.UpdatePackages(context.Background(), 0, nil); err != nil {
		t.Fatalf("unexpected error updating packages: %s", err)
	}

	count, _, err := basestore.ScanFirstInt(db.QueryContext(context.Background(), "SELECT COUNT(*) FROM lsif_packages"))
	if err != nil {
		t.Fatalf("unexpected error checking package count: %s", err)
	}
	if count != 0 {
		t.Errorf("unexpected package count. want=%d have=%d", 0, count)
	}
}

// insertPackages populates the lsif_packages table with the given packages.
func insertPackages(t testing.TB, store Store, packages []shared.Package) {
	for _, pkg := range packages {
		if err := store.UpdatePackages(context.Background(), pkg.DumpID, []precise.Package{
			{
				Scheme:  pkg.Scheme,
				Manager: pkg.Manager,
				Name:    pkg.Name,
				Version: pkg.Version,
			},
		}); err != nil {
			t.Fatalf("unexpected error updating packages: %s", err)
		}
	}
}

// insertPackageReferences populates the lsif_references table with the given package references.
func insertPackageReferences(t testing.TB, store Store, packageReferences []shared.PackageReference) {
	for _, packageReference := range packageReferences {
		if err := store.UpdatePackageReferences(context.Background(), packageReference.DumpID, []precise.PackageReference{
			{
				Package: precise.Package{
					Scheme:  packageReference.Scheme,
					Manager: packageReference.Manager,
					Name:    packageReference.Name,
					Version: packageReference.Version,
				},
			},
		}); err != nil {
			t.Fatalf("unexpected error updating package references: %s", err)
		}
	}
}
