package database

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestSubRepoPermsInsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	db := dbtest.NewDB(t)

	ctx := context.Background()
	s := SubRepoPerms(db, clock)
	prepareSubRepoTestData(ctx, t, s)

	userID := int32(1)
	repoID := api.RepoID(1)
	perms := authz.SubRepoPermissions{
		PathIncludes: []string{"/src/foo/*"},
		PathExcludes: []string{"/src/bar/*"},
	}
	if err := s.Upsert(ctx, userID, repoID, perms); err != nil {
		t.Fatal(err)
	}

	have, err := s.Get(ctx, userID, repoID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(&perms, have); diff != "" {
		t.Fatal(diff)
	}
}

func TestSubRepoPermsUpsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	db := dbtest.NewDB(t)

	ctx := context.Background()
	s := SubRepoPerms(db, clock)
	prepareSubRepoTestData(ctx, t, s)

	userID := int32(1)
	repoID := api.RepoID(1)
	perms := authz.SubRepoPermissions{
		PathIncludes: []string{"/src/foo/*"},
		PathExcludes: []string{"/src/bar/*"},
	}
	// Insert initial data
	if err := s.Upsert(ctx, userID, repoID, perms); err != nil {
		t.Fatal(err)
	}

	// Upsert to change perms
	perms = authz.SubRepoPermissions{
		PathIncludes: []string{"/src/foo_upsert/*"},
		PathExcludes: []string{"/src/bar_upsert/*"},
	}
	if err := s.Upsert(ctx, userID, repoID, perms); err != nil {
		t.Fatal(err)
	}

	have, err := s.Get(ctx, userID, repoID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(&perms, have); diff != "" {
		t.Fatal(diff)
	}
}

func TestSubRepoPermsUpsertWithSpec(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	db := dbtest.NewDB(t)

	ctx := context.Background()
	s := SubRepoPerms(db, clock)
	prepareSubRepoTestData(ctx, t, s)

	userID := int32(1)
	repoID := api.RepoID(1)
	perms := authz.SubRepoPermissions{
		PathIncludes: []string{"/src/foo/*"},
		PathExcludes: []string{"/src/bar/*"},
	}
	spec := api.ExternalRepoSpec{
		ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
		ServiceType: "github",
		ServiceID:   "https://github.com/",
	}
	// Insert initial data
	if err := s.UpsertWithSpec(ctx, userID, spec, perms); err != nil {
		t.Fatal(err)
	}

	// Upsert to change perms
	perms = authz.SubRepoPermissions{
		PathIncludes: []string{"/src/foo_upsert/*"},
		PathExcludes: []string{"/src/bar_upsert/*"},
	}
	if err := s.UpsertWithSpec(ctx, userID, spec, perms); err != nil {
		t.Fatal(err)
	}

	have, err := s.Get(ctx, userID, repoID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(&perms, have); diff != "" {
		t.Fatal(diff)
	}
}

func TestSubRepoPermsGetByUser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	db := dbtest.NewDB(t)

	ctx := context.Background()
	s := SubRepoPerms(db, clock)
	prepareSubRepoTestData(ctx, t, s)

	userID := int32(1)
	perms := authz.SubRepoPermissions{
		PathIncludes: []string{"/src/foo/*"},
		PathExcludes: []string{"/src/bar/*"},
	}
	if err := s.Upsert(ctx, userID, api.RepoID(1), perms); err != nil {
		t.Fatal(err)
	}

	userID = int32(1)
	perms = authz.SubRepoPermissions{
		PathIncludes: []string{"/src/foo2/*"},
		PathExcludes: []string{"/src/bar2/*"},
	}
	if err := s.Upsert(ctx, userID, api.RepoID(2), perms); err != nil {
		t.Fatal(err)
	}

	have, err := s.GetByUser(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	want := map[api.RepoID]authz.SubRepoPermissions{
		1: {
			PathIncludes: []string{"/src/foo/*"},
			PathExcludes: []string{"/src/bar/*"},
		},
		2: {
			PathIncludes: []string{"/src/foo2/*"},
			PathExcludes: []string{"/src/bar2/*"},
		},
	}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}

func prepareSubRepoTestData(ctx context.Context, t *testing.T, s *SubRepoPermsStore) {
	t.Helper()

	// Prepare data
	qs := []*sqlf.Query{
		// ID=1, with newer code host connection sync
		sqlf.Sprintf(`INSERT INTO users(username) VALUES ('alice')`),
		sqlf.Sprintf(`INSERT INTO external_services(id, display_name, kind, config, namespace_user_id, last_sync_at) VALUES(1, 'GitHub #1', 'GITHUB', '{}', 1, NOW() + INTERVAL '10min')`),
		sqlf.Sprintf(`INSERT INTO repo(id, name, external_id, external_service_type, external_service_id) VALUES(1, 'github.com/foo/bar', 'MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==', 'github', 'https://github.com/')`),
		sqlf.Sprintf(`INSERT INTO repo(id, name, external_id, external_service_type, external_service_id) VALUES(2, 'github.com/foo/baz', 'MDEwOlJlcG9zaXRvcnk0MTI4ODcwOB==', 'github', 'https://github.com/')`),
	}
	for _, q := range qs {
		if err := s.Exec(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
}
