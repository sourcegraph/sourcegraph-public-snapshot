package background

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestPackageRepoFiltersBlockOnly(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	s := store.New(&observation.TestContext, db)

	deps := []shared.MinimalPackageRepoRef{
		{Scheme: "npm", Name: "bar", Versions: []string{"2.0.0", "2.0.1", "3.0.0"}},
		{Scheme: "npm", Name: "foo", Versions: []string{"1.0.0"}},
		{Scheme: "npm", Name: "banana", Versions: []string{"2.0.0"}},
		{Scheme: "rust-analyzer", Name: "burger", Versions: []string{"1.0.0", "1.0.1", "1.0.2"}},
	}

	if _, _, err := s.InsertPackageRepoRefs(ctx, deps); err != nil {
		t.Fatal(err)
	}

	bhvr := "BLOCK"
	for _, filter := range []shared.MinimalPackageFilter{
		{
			Behaviour:     &bhvr,
			PackageScheme: "npm",
			NameFilter:    &struct{ PackageGlob string }{PackageGlob: "ba*"},
		}, {
			Behaviour:     &bhvr,
			PackageScheme: "rust-analyzer",
			VersionFilter: &struct {
				PackageName string
				VersionGlob string
			}{
				PackageName: "burger",
				VersionGlob: "1.0.[!1]",
			},
		},
	} {
		if _, err := s.CreatePackageRepoFilter(ctx, filter); err != nil {
			t.Fatal(err)
		}
	}

	job := packagesFilterApplicatorJob{
		store:       s,
		extsvcStore: db.ExternalServices(),
		operations:  newOperations(&observation.TestContext),
	}

	if err := job.handle(ctx); err != nil {
		t.Fatal(err)
	}

	have, count, hasMore, err := s.ListPackageRepoRefs(ctx, store.ListDependencyReposOpts{})
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Errorf("unexpected total count of package repos: want=%d got=%d", 2, count)
	}

	if hasMore {
		t.Error("unexpected more-pages flag set, expected no more pages to follow")
	}

	for i, ref := range have {
		if ref.LastCheckedAt == nil {
			t.Errorf("unexpected nil last_checked_at for package (%s, %s)", ref.Scheme, ref.Name)
		}
		for i, version := range ref.Versions {
			if version.LastCheckedAt == nil {
				t.Errorf("unexpected nil last_checked_at for package version (%s, %s, [%s])", ref.Scheme, ref.Name, version.Version)
			}
			ref.Versions[i].LastCheckedAt = nil
		}
		have[i].LastCheckedAt = nil
	}

	want := []shared.PackageRepoReference{
		{ID: 3, Scheme: "rust-analyzer", Name: "burger", Versions: []shared.PackageRepoRefVersion{{ID: 6, PackageRefID: 3, Version: "1.0.1"}}},
		{ID: 4, Scheme: "npm", Name: "foo", Versions: []shared.PackageRepoRefVersion{{ID: 8, PackageRefID: 4, Version: "1.0.0"}}},
	}
	if diff := cmp.Diff(want, have); diff != "" {
		t.Errorf("mismatch (-want, +got): %s", diff)
	}
}
