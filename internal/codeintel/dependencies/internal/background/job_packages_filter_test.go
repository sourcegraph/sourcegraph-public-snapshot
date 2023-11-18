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
	db := database.NewDB(logger, dbtest.NewDB(t))
	s := store.New(&observation.TestContext, db)

	deps := []shared.MinimalPackageRepoRef{
		{Scheme: "npm", Name: "bar", Versions: []shared.MinimalPackageRepoRefVersion{{Version: "2.0.0"}, {Version: "2.0.1"}, {Version: "3.0.0"}}},
		{Scheme: "npm", Name: "foo", Versions: []shared.MinimalPackageRepoRefVersion{{Version: "1.0.0"}}},
		{Scheme: "npm", Name: "banana", Versions: []shared.MinimalPackageRepoRefVersion{{Version: "2.0.0"}}},
		{Scheme: "rust-analyzer", Name: "burger", Versions: []shared.MinimalPackageRepoRefVersion{{Version: "1.0.0"}, {Version: "1.0.1"}, {Version: "1.0.2"}}},
		// make sure filters only apply to their respective scheme
		{Scheme: "semanticdb", Name: "burger", Versions: []shared.MinimalPackageRepoRefVersion{{Version: "1.0.3"}}},
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

	if count != 3 {
		t.Errorf("unexpected total count of package repos: want=%d got=%d", 3, count)
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
		{ID: 4, Scheme: "semanticdb", Name: "burger", Versions: []shared.PackageRepoRefVersion{{ID: 8, PackageRefID: 4, Version: "1.0.3"}}},
		{ID: 5, Scheme: "npm", Name: "foo", Versions: []shared.PackageRepoRefVersion{{ID: 9, PackageRefID: 5, Version: "1.0.0"}}},
	}
	if diff := cmp.Diff(want, have); diff != "" {
		t.Errorf("mismatch (-want, +got): %s", diff)
	}
}

func TestPackageRepoFiltersBlockAllow(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	s := store.New(&observation.TestContext, db)

	deps := []shared.MinimalPackageRepoRef{
		{Scheme: "npm", Name: "bar", Versions: []shared.MinimalPackageRepoRefVersion{{Version: "2.0.0"}, {Version: "2.0.1"}, {Version: "3.0.0"}}},
		{Scheme: "npm", Name: "foo", Versions: []shared.MinimalPackageRepoRefVersion{{Version: "1.0.0"}}},
		{Scheme: "npm", Name: "banana", Versions: []shared.MinimalPackageRepoRefVersion{{Version: "2.0.0"}}},
		{Scheme: "rust-analyzer", Name: "burger", Versions: []shared.MinimalPackageRepoRefVersion{{Version: "1.0.0"}, {Version: "1.0.1"}, {Version: "1.0.2"}}},
		{Scheme: "rust-analyzer", Name: "frogger", Versions: []shared.MinimalPackageRepoRefVersion{{Version: "4.1.2"}, {Version: "3.0.0"}}},
		{Scheme: "semanticdb", Name: "burger", Versions: []shared.MinimalPackageRepoRefVersion{{Version: "1.0.3"}}},
	}

	if _, _, err := s.InsertPackageRepoRefs(ctx, deps); err != nil {
		t.Fatal(err)
	}

	block := "BLOCK"
	allow := "ALLOW"
	for _, filter := range []shared.MinimalPackageFilter{
		{
			Behaviour:     &block,
			PackageScheme: "npm",
			NameFilter:    &struct{ PackageGlob string }{PackageGlob: "ba*"},
		},
		{
			Behaviour:     &allow,
			PackageScheme: "rust-analyzer",
			VersionFilter: &struct {
				PackageName string
				VersionGlob string
			}{
				PackageName: "burger",
				VersionGlob: "1.0.[!1]",
			},
		},
		{
			Behaviour:     &allow,
			PackageScheme: "rust-analyzer",
			VersionFilter: &struct {
				PackageName string
				VersionGlob string
			}{
				PackageName: "frogger",
				VersionGlob: "3*",
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

	if count != 4 {
		t.Errorf("unexpected total count of package repos: want=%d got=%d", 4, count)
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
		{ID: 3, Scheme: "rust-analyzer", Name: "burger", Versions: []shared.PackageRepoRefVersion{{ID: 5, PackageRefID: 3, Version: "1.0.0"}, {ID: 7, PackageRefID: 3, Version: "1.0.2"}}},
		{ID: 4, Scheme: "semanticdb", Name: "burger", Versions: []shared.PackageRepoRefVersion{{ID: 8, PackageRefID: 4, Version: "1.0.3"}}},
		{ID: 5, Scheme: "npm", Name: "foo", Versions: []shared.PackageRepoRefVersion{{ID: 9, PackageRefID: 5, Version: "1.0.0"}}},
		{ID: 6, Scheme: "rust-analyzer", Name: "frogger", Versions: []shared.PackageRepoRefVersion{{ID: 10, PackageRefID: 6, Version: "3.0.0"}}},
	}
	if diff := cmp.Diff(want, have); diff != "" {
		t.Errorf("mismatch (-want, +got): %s", diff)
	}
}
