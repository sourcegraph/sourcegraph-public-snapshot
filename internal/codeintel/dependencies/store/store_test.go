package store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestUpsertDependencyRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := database.NewDB(dbtest.NewDB(t))
	store := testStore(db)

	for _, dep := range []struct {
		reposource.PackageDependency
		isNew bool
	}{
		{mustParseNPMDependency(t, "bar@2.0.0"), true},
		{mustParseNPMDependency(t, "bar@2.0.0"), false},
		{mustParseNPMDependency(t, "bar@3.0.0"), true},
		{mustParseNPMDependency(t, "foo@1.0.0"), true},
		{mustParseNPMDependency(t, "foo@1.0.0"), false},
		{mustParseNPMDependency(t, "foo@2.0.0"), true},
	} {
		isNew, err := store.UpsertDependencyRepo(ctx, dep)
		if err != nil {
			t.Fatal(err)
		}

		if have, want := isNew, dep.isNew; have != want {
			t.Fatalf("%s: want isNew=%t, have %t", dep.PackageManagerSyntax(), want, have)
		}
	}

	have, err := store.ListDependencyRepos(ctx, ListDependencyReposOpts{
		Scheme: NPMPackagesScheme,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []DependencyRepo{
		{ID: 6, Scheme: "npm", Name: "foo", Version: "2.0.0"},
		{ID: 4, Scheme: "npm", Name: "foo", Version: "1.0.0"},
		{ID: 3, Scheme: "npm", Name: "bar", Version: "3.0.0"},
		{ID: 1, Scheme: "npm", Name: "bar", Version: "2.0.0"},
	}

	opt := cmpopts.IgnoreFields(DependencyRepo{}, "ID")
	if diff := cmp.Diff(have, want, opt); diff != "" {
		t.Fatalf("mismatch (-have, +want): %s", diff)
	}
}

func mustParseNPMDependency(t testing.TB, dep string) reposource.PackageDependency {
	t.Helper()

	d, err := reposource.ParseNPMDependency(dep)
	if err != nil {
		t.Fatal(err)
	}

	return d
}

func testStore(db dbutil.DB) *Store {
	return newStore(db, &observation.TestContext)
}
