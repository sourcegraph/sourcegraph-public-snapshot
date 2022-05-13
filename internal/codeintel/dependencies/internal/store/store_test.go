package store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestLockfileDependencies(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := database.NewDB(dbtest.NewDB(t))
	store := TestStore(db)

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (name) VALUES ('foo')`); err != nil {
		t.Fatalf(err.Error())
	}

	packageA := shared.TestPackageDependencyLiteral(api.RepoName("A"), "1", "2", "3", "4")
	packageB := shared.TestPackageDependencyLiteral(api.RepoName("B"), "2", "3", "4", "5")
	packageC := shared.TestPackageDependencyLiteral(api.RepoName("C"), "3", "4", "5", "6")
	packageD := shared.TestPackageDependencyLiteral(api.RepoName("D"), "4", "5", "6", "7")
	packageE := shared.TestPackageDependencyLiteral(api.RepoName("E"), "5", "6", "7", "8")
	packageF := shared.TestPackageDependencyLiteral(api.RepoName("F"), "6", "7", "8", "9")

	commits := map[string][]shared.PackageDependency{
		"cafebabe": {packageA, packageB, packageC},
		"deadbeef": {packageA, packageB, packageD, packageE},
		"deadc0de": {packageB, packageF},
		"deadd00d": nil,
	}

	for commit, deps := range commits {
		if err := store.UpsertLockfileDependencies(ctx, "foo", commit, deps); err != nil {
			t.Fatalf("unexpected error upserting lockfile dependencies: %s", err)
		}
	}

	// Update twice to show idempotency
	for commit, expected := range commits {
		if err := store.UpsertLockfileDependencies(ctx, "foo", commit, expected); err != nil {
			t.Fatalf("unexpected error upserting lockfile dependencies: %s", err)
		}
	}

	for commit, expectedDeps := range commits {
		deps, found, err := store.LockfileDependencies(ctx, "foo", commit)
		if err != nil {
			t.Fatalf("unexpected error querying lockfile dependencies of %s: %s", commit, err)
		}
		if !found {
			t.Fatalf("expected dependencies to be cached for %s", commit)
		}

		if diff := cmp.Diff(expectedDeps, deps); diff != "" {
			t.Fatalf("unexpected dependencies for commit %s (-have, +want): %s", commit, diff)
		}
	}

	missingCommit := "d00dd00d"
	_, found, err := store.LockfileDependencies(ctx, "foo", missingCommit)
	if err != nil {
		t.Fatalf("unexpected error querying lockfile dependencies: %s", err)
	} else if found {
		t.Fatalf("expected no dependencies to be cached for %s", missingCommit)
	}
}

func TestUpsertDependencyRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := database.NewDB(dbtest.NewDB(t))
	store := TestStore(db)

	batches := [][]shared.Repo{
		{
			// Test same-set flushes
			shared.Repo{Scheme: "npm", Name: "bar", Version: "2.0.0"}, // id=1
			shared.Repo{Scheme: "npm", Name: "bar", Version: "2.0.0"}, // id=2, duplicate
		},
		{
			shared.Repo{Scheme: "npm", Name: "bar", Version: "3.0.0"}, // id=3
			shared.Repo{Scheme: "npm", Name: "foo", Version: "1.0.0"}, // id=4
		},
		{
			// Test different-set flushes
			shared.Repo{Scheme: "npm", Name: "foo", Version: "1.0.0"}, // id=5, duplicate
			shared.Repo{Scheme: "npm", Name: "foo", Version: "2.0.0"}, // id=6
		},
	}

	var allNewDeps []shared.Repo
	for _, batch := range batches {
		newDeps, err := store.UpsertDependencyRepos(ctx, batch)
		if err != nil {
			t.Fatal(err)
		}

		allNewDeps = append(allNewDeps, newDeps...)
	}

	want := []shared.Repo{
		{ID: 1, Scheme: "npm", Name: "bar", Version: "2.0.0"},
		{ID: 3, Scheme: "npm", Name: "bar", Version: "3.0.0"},
		{ID: 4, Scheme: "npm", Name: "foo", Version: "1.0.0"},
		{ID: 6, Scheme: "npm", Name: "foo", Version: "2.0.0"},
	}
	if diff := cmp.Diff(allNewDeps, want); diff != "" {
		t.Fatalf("mismatch (-have, +want): %s", diff)
	}

	have, err := store.ListDependencyRepos(ctx, ListDependencyReposOpts{
		Scheme: shared.NpmPackagesScheme,
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatalf("mismatch (-have, +want): %s", diff)
	}
}

func TestDeleteDependencyReposByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := database.NewDB(dbtest.NewDB(t))
	store := TestStore(db)

	repos := []shared.Repo{
		// Test same-set flushes
		{ID: 1, Scheme: "npm", Name: "bar", Version: "2.0.0"},
		{ID: 2, Scheme: "npm", Name: "bar", Version: "3.0.0"}, // deleted
		{ID: 3, Scheme: "npm", Name: "foo", Version: "1.0.0"}, // deleted
		{ID: 4, Scheme: "npm", Name: "foo", Version: "2.0.0"},
	}

	if _, err := store.UpsertDependencyRepos(ctx, repos); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteDependencyReposByID(ctx, 2, 3); err != nil {
		t.Fatalf(err.Error())
	}

	have, err := store.ListDependencyRepos(ctx, ListDependencyReposOpts{
		Scheme: shared.NpmPackagesScheme,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []shared.Repo{
		{ID: 1, Scheme: "npm", Name: "bar", Version: "2.0.0"},
		{ID: 4, Scheme: "npm", Name: "foo", Version: "2.0.0"},
	}
	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatalf("mismatch (-have, +want): %s", diff)
	}
}
