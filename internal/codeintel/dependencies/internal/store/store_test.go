package store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestUpsertDependencyRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	batches := [][]shared.MinimalPackageRepoRef{
		{
			// Test same-set flushes
			shared.MinimalPackageRepoRef{Scheme: "npm", Name: "bar", Versions: []string{"2.0.0"}},
			shared.MinimalPackageRepoRef{Scheme: "npm", Name: "bar", Versions: []string{"2.0.0"}},
		},
		{
			shared.MinimalPackageRepoRef{Scheme: "npm", Name: "bar", Versions: []string{"3.0.0"}}, // id=3
			shared.MinimalPackageRepoRef{Scheme: "npm", Name: "foo", Versions: []string{"1.0.0"}}, // id=4
		},
		{
			// Test different-set flushes
			shared.MinimalPackageRepoRef{Scheme: "npm", Name: "foo", Versions: []string{"1.0.0", "2.0.0"}},
		},
	}

	var allNewDeps []shared.PackageRepoReference
	var allNewVersions []shared.PackageRepoRefVersion
	for _, batch := range batches {
		newDeps, newVersions, err := store.InsertPackageRepoRefs(ctx, batch)
		if err != nil {
			t.Fatal(err)
		}

		allNewDeps = append(allNewDeps, newDeps...)
		allNewVersions = append(allNewVersions, newVersions...)
	}

	want := []shared.PackageRepoReference{
		{ID: 1, Scheme: "npm", Name: "bar"},
		{ID: 2, Scheme: "npm", Name: "foo"},
	}
	if diff := cmp.Diff(want, allNewDeps); diff != "" {
		t.Fatalf("mismatch (-want, +got): %s", diff)
	}

	wantV := []shared.PackageRepoRefVersion{
		{ID: 1, PackageRefID: 1, Version: "2.0.0"},
		{ID: 2, PackageRefID: 1, Version: "3.0.0"},
		{ID: 3, PackageRefID: 2, Version: "1.0.0"},
		{ID: 4, PackageRefID: 2, Version: "2.0.0"},
	}
	if diff := cmp.Diff(wantV, allNewVersions); diff != "" {
		t.Fatalf("mismatch (-want, +got): %s", diff)
	}

	have, _, err := store.ListPackageRepoRefs(ctx, ListDependencyReposOpts{
		Scheme: shared.NpmPackagesScheme,
	})
	if err != nil {
		t.Fatal(err)
	}

	want[0].Versions = []shared.PackageRepoRefVersion{{ID: 1, PackageRefID: 1, Version: "2.0.0"}, {ID: 2, PackageRefID: 1, Version: "3.0.0"}}
	want[1].Versions = []shared.PackageRepoRefVersion{{ID: 3, PackageRefID: 2, Version: "1.0.0"}, {ID: 4, PackageRefID: 2, Version: "2.0.0"}}
	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatalf("mismatch (-want, +got): %s", diff)
	}
}

func TestListPackageRepoRefs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	batches := []shared.MinimalPackageRepoRef{
		{Scheme: "npm", Name: "bar", Versions: []string{"2.0.0"}},    // id=1
		{Scheme: "npm", Name: "foo", Versions: []string{"1.0.0"}},    // id=2
		{Scheme: "npm", Name: "bar", Versions: []string{"2.0.1"}},    // id=3
		{Scheme: "npm", Name: "foo", Versions: []string{"1.0.0"}},    // id=4
		{Scheme: "npm", Name: "bar", Versions: []string{"3.0.0"}},    // id=5
		{Scheme: "npm", Name: "banana", Versions: []string{"2.0.0"}}, // id=6
		{Scheme: "npm", Name: "turtle", Versions: []string{"4.2.0"}}, // id=7
	}

	if _, _, err := store.InsertPackageRepoRefs(ctx, batches); err != nil {
		t.Fatal(err)
	}

	var lastID int
	for _, test := range [][]shared.PackageRepoReference{
		{{Scheme: "npm", Name: "banana"}, {Scheme: "npm", Name: "bar"}, {Scheme: "npm", Name: "foo"}},
		{{Scheme: "npm", Name: "turtle"}},
	} {
		depRepos, _, err := store.ListPackageRepoRefs(ctx, ListDependencyReposOpts{
			Scheme: "npm",
			After:  lastID,
			Limit:  3,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lastID = depRepos[len(depRepos)-1].ID

		for i := range depRepos {
			depRepos[i].ID = 0
			depRepos[i].Versions = nil
		}

		if diff := cmp.Diff(test, depRepos); diff != "" {
			t.Errorf("mismatch (-want, +got): %s", diff)
		}
	}
}

func TestDeletePackageRepoRefsByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	repos := []shared.MinimalPackageRepoRef{
		// Test same-set flushes
		{Scheme: "npm", Name: "bar", Versions: []string{"2.0.0"}},
		{Scheme: "npm", Name: "bar", Versions: []string{"3.0.0"}}, // deleted
		{Scheme: "npm", Name: "foo", Versions: []string{"1.0.0"}}, // deleted
		{Scheme: "npm", Name: "foo", Versions: []string{"2.0.0"}},
		{Scheme: "npm", Name: "banan", Versions: []string{"4.2.0"}}, // deleted
	}

	if _, _, err := store.InsertPackageRepoRefs(ctx, repos); err != nil {
		t.Fatal(err)
	}
	if err := store.DeletePackageRepoRefsByID(ctx, 1); err != nil {
		t.Fatal(err)
	}

	if err := store.DeletePackageRepoRefVersionsByID(ctx, 3, 4); err != nil {
		t.Fatal(err)
	}

	have, _, err := store.ListPackageRepoRefs(ctx, ListDependencyReposOpts{
		Scheme: shared.NpmPackagesScheme,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []shared.PackageRepoReference{
		{ID: 2, Scheme: "npm", Name: "bar", Versions: []shared.PackageRepoRefVersion{{ID: 2, PackageRefID: 2, Version: "2.0.0"}}},
		{ID: 3, Scheme: "npm", Name: "foo", Versions: []shared.PackageRepoRefVersion{{ID: 5, PackageRefID: 3, Version: "2.0.0"}}},
	}
	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatalf("mismatch (-want, +got): %s", diff)
	}
}
