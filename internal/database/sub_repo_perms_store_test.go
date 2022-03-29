package database

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func TestSubRepoPermsInsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	db := dbtest.NewDB(t)

	ctx := context.Background()
	prepareSubRepoTestData(ctx, t, db)
	s := SubRepoPerms(db)

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
	prepareSubRepoTestData(ctx, t, db)
	s := SubRepoPerms(db)

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
	prepareSubRepoTestData(ctx, t, db)
	s := SubRepoPerms(db)

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
	s := SubRepoPerms(db)
	prepareSubRepoTestData(ctx, t, db)

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

	want := map[api.RepoName]authz.SubRepoPermissions{
		"github.com/foo/bar": {
			PathIncludes: []string{"/src/foo/*"},
			PathExcludes: []string{"/src/bar/*"},
		},
		"github.com/foo/baz": {
			PathIncludes: []string{"/src/foo2/*"},
			PathExcludes: []string{"/src/bar2/*"},
		},
	}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}

func TestSubRepoPermsSupportedForRepoId(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	db := dbtest.NewDB(t)

	ctx := context.Background()
	s := SubRepoPerms(db)
	prepareSubRepoTestData(ctx, t, db)

	testSubRepoNotSupportedForRepo(ctx, t, s, 3, "perforce1", "Repo is not private, therefore sub-repo perms are not supported")

	testSubRepoSupportedForRepo(ctx, t, s, 4, "perforce2", "Repo is private, therefore sub-repo perms are supported")

	testSubRepoNotSupportedForRepo(ctx, t, s, 5, "github.com/foo/qux", "Repo is not perforce, therefore sub-repo perms are not supported")
}

func testSubRepoNotSupportedForRepo(ctx context.Context, t *testing.T, s SubRepoPermsStore, repoID api.RepoID, repoName api.RepoName, errMsg string) {
	t.Helper()
	exists, err := s.RepoIdSupported(ctx, repoID)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal(errMsg)
	}
	exists, err = s.RepoSupported(ctx, repoName)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal(errMsg)
	}
}

func testSubRepoSupportedForRepo(ctx context.Context, t *testing.T, s SubRepoPermsStore, repoID api.RepoID, repoName api.RepoName, errMsg string) {
	t.Helper()
	exists, err := s.RepoIdSupported(ctx, repoID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal(errMsg)
	}
	exists, err = s.RepoSupported(ctx, repoName)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal(errMsg)
	}
}

func prepareSubRepoTestData(ctx context.Context, t *testing.T, db dbutil.DB) {
	t.Helper()

	// Prepare data
	qs := []string{
		`INSERT INTO users(username) VALUES ('alice')`,

		`INSERT INTO external_services(id, display_name, kind, config, namespace_user_id, last_sync_at) VALUES(1, 'GitHub #1', 'GITHUB', '{}', 1, NOW() + INTERVAL '10min')`,
		`INSERT INTO external_services(id, display_name, kind, config, namespace_user_id, last_sync_at) VALUES(2, 'Perforce #1', 'PERFORCE', '{}', 1, NOW() + INTERVAL '10min')`,

		`INSERT INTO repo(id, name, external_id, external_service_type, external_service_id) VALUES(1, 'github.com/foo/bar', 'MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==', 'github', 'https://github.com/')`,
		`INSERT INTO repo(id, name, external_id, external_service_type, external_service_id) VALUES(2, 'github.com/foo/baz', 'MDEwOlJlcG9zaXRvcnk0MTI4ODcwOB==', 'github', 'https://github.com/')`,
		`INSERT INTO repo(id, name, external_id, external_service_type, external_service_id) VALUES(3, 'perforce1', 'MDEwOlJlcG9zaXRvcnk0MTI4ODcwOB==', 'perforce', 'https://perforce.com/')`,
		`INSERT INTO repo(id, name, external_id, external_service_type, external_service_id, private) VALUES(4, 'perforce2', 'MDEwOlJlcG9zaXRvcnk0MTI4ODcwOB==', 'perforce', 'https://perforce.com/2', 'true')`,
		`INSERT INTO repo(id, name, external_id, external_service_type, external_service_id, private) VALUES(5, 'github.com/foo/qux', 'MDEwOlJlcG9zaXRvcnk0MTI4ODcwOC==', 'github', 'https://github.com/', 'true')`,

		`INSERT INTO external_service_repos(repo_id, external_service_id, clone_url) VALUES(1, 1, 'cloneURL')`,
		`INSERT INTO external_service_repos(repo_id, external_service_id, clone_url) VALUES(2, 1, 'cloneURL')`,
		`INSERT INTO external_service_repos(repo_id, external_service_id, clone_url) VALUES(3, 2, 'cloneURL')`,
	}
	for _, q := range qs {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
}
