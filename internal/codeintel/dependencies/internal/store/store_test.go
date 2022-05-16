package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
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

func TestSelectRepoRevisionsToResolve(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := database.NewDB(dbtest.NewDB(t))
	store := TestStore(db)

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (name) VALUES ('repo-1')`); err != nil {
		t.Fatalf(err.Error())
	}

	packageA := shared.TestPackageDependencyLiteral(api.RepoName("A"), "v1", "2", "3", "4")
	packageB := shared.TestPackageDependencyLiteral(api.RepoName("B"), "v2", "3", "4", "5")
	packageC := shared.TestPackageDependencyLiteral(api.RepoName("C"), "v3", "4", "5", "6")
	packageD := shared.TestPackageDependencyLiteral(api.RepoName("D"), "v4", "5", "6", "7")
	packageE := shared.TestPackageDependencyLiteral(api.RepoName("E"), "v5", "6", "7", "8")

	commit := "d34df00d"
	repoName := "repo-1"
	packages := []shared.PackageDependency{packageA, packageB, packageC, packageD, packageE}

	if err := store.UpsertLockfileDependencies(ctx, repoName, commit, packages); err != nil {
		t.Fatalf("unexpected error upserting lockfile dependencies: %s", err)
	}

	now := timeutil.Now()

	selected, err := store.selectRepoRevisionsToResolve(ctx, 3, 24*time.Hour, now)
	if err != nil {
		t.Fatalf("unexpected error selecting repo revisions to resolve: %s", err)
	}

	expectedRepoRevisions := map[string][]string{
		"A": {"v1"},
		"B": {"v2"},
		"C": {"v3"},
	}
	if diff := cmp.Diff(selected, expectedRepoRevisions); diff != "" {
		t.Errorf("unexpected sourced commits (-want +got):\n%s", diff)
	}

	selected, err = store.selectRepoRevisionsToResolve(ctx, 3, 24*time.Hour, now)
	if err != nil {
		t.Fatalf("unexpected error selecting repo revisions to resolve: %s", err)
	}

	expectedRepoRevisions = map[string][]string{
		"D": {"v4"},
		"E": {"v5"},
	}
	if diff := cmp.Diff(selected, expectedRepoRevisions); diff != "" {
		t.Errorf("unexpected sourced commits (-want +got):\n%s", diff)
	}

	// Run it again, but all should be resolved in timeframe
	selected, err = store.selectRepoRevisionsToResolve(ctx, 6, 24*time.Hour, now)
	if err != nil {
		t.Fatalf("unexpected error selecting repo revisions to resolve: %s", err)
	}

	expectedRepoRevisions = map[string][]string{}
	if diff := cmp.Diff(selected, expectedRepoRevisions); diff != "" {
		t.Errorf("unexpected sourced commits (-want +got):\n%s", diff)
	}

	// Run it again, but in the future, all should be resolved in timeframe
	now = now.Add(24 * time.Hour)

	selected, err = store.selectRepoRevisionsToResolve(ctx, 6, 24*time.Hour, now)
	if err != nil {
		t.Fatalf("unexpected error selecting repo revisions to resolve: %s", err)
	}

	expectedRepoRevisions = map[string][]string{
		"A": {"v1"},
		"B": {"v2"},
		"C": {"v3"},
		"D": {"v4"},
		"E": {"v5"},
	}
	if diff := cmp.Diff(selected, expectedRepoRevisions); diff != "" {
		t.Errorf("unexpected sourced commits (-want +got):\n%s", diff)
	}
}

func TestUpdateResolvedRevisions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := database.NewDB(dbtest.NewDB(t))
	store := TestStore(db)

	for _, repo := range []string{"repo-1", "repo-2", "repo-3"} {
		if err := store.Exec(ctx, sqlf.Sprintf(`INSERT INTO repo (name) VALUES (%s)`, repo)); err != nil {
			t.Fatalf(err.Error())
		}
	}

	packageA := shared.TestPackageDependencyLiteral(api.RepoName("A"), "v1", "2", "3", "4")
	packageB := shared.TestPackageDependencyLiteral(api.RepoName("repo-2"), "v2", "3", "4", "5")
	packageC := shared.TestPackageDependencyLiteral(api.RepoName("repo-3"), "v3", "4", "5", "6")
	packageD := shared.TestPackageDependencyLiteral(api.RepoName("D"), "v4", "5", "6", "7")

	commit := "d34df00d"
	repoName := "repo-1"
	packages := []shared.PackageDependency{packageA, packageB, packageC, packageD}

	if err := store.UpsertLockfileDependencies(ctx, repoName, commit, packages); err != nil {
		t.Fatalf("unexpected error upserting lockfile dependencies: %s", err)
	}

	resolvedRevisions := map[string]map[string]string{
		"repo-2": {"v2": "d34df00d"},
		"repo-3": {"v3": "d34db33f"},
	}

	if err := store.UpdateResolvedRevisions(ctx, resolvedRevisions); err != nil {
		t.Fatalf("unexpected error updating resolved revisions: %s", err)
	}

	for repoName, resolvedRevs := range resolvedRevisions {
		for revspec, commit := range resolvedRevs {
			q := sqlf.Sprintf(`
			SELECT 1 FROM codeintel_lockfile_references
			WHERE
				repository_id = (SELECT id FROM repo WHERE name = %s)
			AND
				revspec = %s
			AND
				commit_bytea = %s
			`,
				repoName,
				revspec,
				dbutil.CommitBytea(commit),
			)

			_, ok, err := basestore.ScanFirstInt(store.Query(ctx, q))
			if err != nil {
				t.Fatalf("failed to query resolved revisions for repo %s: %s", repoName, err)
			}
			if !ok {
				t.Fatalf("revspec %q for repo %q was not updated to commit %s", revspec, repoName, commit)
			}
		}
	}
}
