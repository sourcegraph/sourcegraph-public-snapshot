package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSubRepoPermsInsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))

	ctx := context.Background()
	prepareSubRepoTestData(ctx, t, db)
	s := db.SubRepoPerms()

	userID := int32(1)
	repoID := api.RepoID(1)
	perms := authz.SubRepoPermissions{
		Paths: []string{"/src/foo/*", "-/src/bar/*"},
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

func TestSubRepoPermsDeleteByUser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))

	ctx := context.Background()
	prepareSubRepoTestData(ctx, t, db)
	s := db.SubRepoPerms()

	userID := int32(1)
	repoID := api.RepoID(1)
	perms := authz.SubRepoPermissions{
		Paths: []string{"/src/foo/*", "-/src/bar/*"},
	}
	if err := s.Upsert(ctx, userID, repoID, perms); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteByUser(ctx, userID); err != nil {
		t.Fatal(err)
	}
	have, err := s.Get(ctx, userID, repoID)
	if err != nil {
		t.Fatal(err)
	}

	want := authz.SubRepoPermissions{}
	if diff := cmp.Diff(&want, have); diff != "" {
		t.Fatal(diff)
	}
}

func TestSubRepoPermsUpsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))

	ctx := context.Background()
	prepareSubRepoTestData(ctx, t, db)
	s := db.SubRepoPerms()

	userID := int32(1)
	repoID := api.RepoID(1)
	perms := authz.SubRepoPermissions{
		Paths: []string{"/src/foo/*", "-/src/bar/*"},
	}
	// Insert initial data
	if err := s.Upsert(ctx, userID, repoID, perms); err != nil {
		t.Fatal(err)
	}

	// Upsert to change perms
	perms = authz.SubRepoPermissions{
		Paths: []string{"/src/foo_upsert/*", "-/src/bar_upsert/*"},
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

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))

	ctx := context.Background()
	prepareSubRepoTestData(ctx, t, db)
	s := db.SubRepoPerms()

	userID := int32(1)
	repoID := api.RepoID(1)
	perms := authz.SubRepoPermissions{
		Paths: []string{"/src/foo/*", "-/src/bar/*"},
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
		Paths: []string{"/src/foo_upsert/*", "-/src/bar_upsert/*"},
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

	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthzEnforceForSiteAdmins: true}})
	t.Cleanup(func() { conf.Mock(nil) })

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))

	ctx := context.Background()
	s := db.SubRepoPerms()
	prepareSubRepoTestData(ctx, t, db)

	userID := int32(1)
	perms := authz.SubRepoPermissions{
		Paths: []string{"/src/foo/*", "-/src/bar/*"},
	}
	if err := s.Upsert(ctx, userID, api.RepoID(1), perms); err != nil {
		t.Fatal(err)
	}

	userID = int32(1)
	perms = authz.SubRepoPermissions{
		Paths: []string{"/src/foo2/*", "-/src/bar2/*"},
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
			Paths: []string{"/src/foo/*", "-/src/bar/*"},
		},
		"github.com/foo/baz": {
			Paths: []string{"/src/foo2/*", "-/src/bar2/*"},
		},
	}
	assert.Equal(t, want, have)

	// Check all combinations of site admin / AuthzEnforceForSiteAdmins
	for _, tc := range []struct {
		siteAdmin           bool
		enforceForSiteAdmin bool
		wantRows            bool
	}{
		{siteAdmin: true, enforceForSiteAdmin: true, wantRows: true},
		{siteAdmin: false, enforceForSiteAdmin: false, wantRows: true},
		{siteAdmin: true, enforceForSiteAdmin: false, wantRows: false},
		{siteAdmin: false, enforceForSiteAdmin: true, wantRows: true},
	} {
		t.Run(fmt.Sprintf("SiteAdmin:%t-Enforce:%t", tc.siteAdmin, tc.enforceForSiteAdmin), func(t *testing.T) {
			conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthzEnforceForSiteAdmins: tc.enforceForSiteAdmin}})
			result, err := db.ExecContext(ctx, "UPDATE users SET site_admin = $1 WHERE id = $2", tc.siteAdmin, userID)
			if err != nil {
				t.Fatal(err)
			}
			affected, err := result.RowsAffected()
			if err != nil {
				t.Fatal(err)
			}
			if affected != 1 {
				t.Fatalf("Wanted 1 row affected, got %d", affected)
			}

			have, err = s.GetByUser(ctx, userID)
			if err != nil {
				t.Fatal(err)
			}
			if tc.wantRows {
				assert.NotEmpty(t, have)
			} else {
				assert.Empty(t, have)
			}
		})
	}
}

func TestSubRepoPermsGetByUserAndService(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(t))

	ctx := context.Background()
	s := db.SubRepoPerms()
	prepareSubRepoTestData(ctx, t, db)

	userID := int32(1)
	perms := authz.SubRepoPermissions{
		Paths: []string{"/src/foo/*", "-/src/bar/*"},
	}
	if err := s.Upsert(ctx, userID, api.RepoID(1), perms); err != nil {
		t.Fatal(err)
	}

	userID = int32(1)
	perms = authz.SubRepoPermissions{
		Paths: []string{"/src/foo2/*", "-/src/bar2/*"},
	}
	if err := s.Upsert(ctx, userID, api.RepoID(2), perms); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name        string
		userID      int32
		serviceType string
		serviceID   string
		want        map[api.ExternalRepoSpec]authz.SubRepoPermissions
	}{
		{
			name:        "Unknown service",
			userID:      userID,
			serviceType: "abc",
			serviceID:   "xyz",
			want:        map[api.ExternalRepoSpec]authz.SubRepoPermissions{},
		},
		{
			name:        "Known service",
			userID:      userID,
			serviceType: "github",
			serviceID:   "https://github.com/",
			want: map[api.ExternalRepoSpec]authz.SubRepoPermissions{
				{
					ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
					ServiceType: "github",
					ServiceID:   "https://github.com/",
				}: {
					Paths: []string{"/src/foo/*", "-/src/bar/*"},
				},
				{
					ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOB==",
					ServiceType: "github",
					ServiceID:   "https://github.com/",
				}: {
					Paths: []string{"/src/foo2/*", "-/src/bar2/*"},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			have, err := s.GetByUserAndService(ctx, userID, tc.serviceType, tc.serviceID)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, have); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestSubRepoPermsSupportedForRepoId(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))

	ctx := context.Background()
	s := db.SubRepoPerms()
	prepareSubRepoTestData(ctx, t, db)

	testSubRepoNotSupportedForRepo(ctx, t, s, 3, "perforce1", "Repo is not private, therefore sub-repo perms are not supported")
	testSubRepoSupportedForRepo(ctx, t, s, 4, "perforce2", "Repo is private, therefore sub-repo perms are supported")
	testSubRepoNotSupportedForRepo(ctx, t, s, 5, "github.com/foo/qux", "Repo is not perforce, therefore sub-repo perms are not supported")
}

func testSubRepoNotSupportedForRepo(ctx context.Context, t *testing.T, s SubRepoPermsStore, repoID api.RepoID, repoName api.RepoName, errMsg string) {
	t.Helper()
	exists, err := s.RepoIDSupported(ctx, repoID)
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
	exists, err := s.RepoIDSupported(ctx, repoID)
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
		`INSERT INTO users(username ) VALUES ('alice')`,

		`INSERT INTO external_services(id, display_name, kind, config,  last_sync_at) VALUES(1, 'GitHub #1', 'GITHUB', '{}', NOW() + INTERVAL '10min')`,
		`INSERT INTO external_services(id, display_name, kind, config,  last_sync_at) VALUES(2, 'Perforce #1', 'PERFORCE', '{}', NOW() + INTERVAL '10min')`,

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

func TestUpsertWithIPsDisallowsEmptyIPs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	prepareSubRepoTestData(ctx, t, db)

	s := db.SubRepoPerms()

	userID := int32(1)
	repoID := api.RepoID(1)

	perms := authz.SubRepoPermissionsWithIPs{
		Paths: []authz.PathWithIP{
			{Path: "/src/foo/*", IP: ""},
			{Path: "-/src/bar/*", IP: ""},
		},
	}

	if err := s.UpsertWithIPs(ctx, userID, repoID, perms); err == nil {
		t.Fatal("UpsertWithIPs should have failed with empty IPs")
	}
}

func TestSubRepoPermsStore_UpsertWithIps(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	newStore := func(t *testing.T) SubRepoPermsStore {
		t.Helper()

		logger := logtest.Scoped(t)
		db := NewDB(logger, dbtest.NewDB(t))

		ctx := context.Background()
		prepareSubRepoTestData(ctx, t, db)
		s := db.SubRepoPerms()

		return s
	}

	userID := int32(1)
	repoID := api.RepoID(1)

	ctx := context.Background()

	for _, test := range []struct {
		name      string
		storeFunc func(t *testing.T) SubRepoPermsStore
	}{
		{
			name:      "empty store",
			storeFunc: newStore,
		},
		{
			name: "store with initial normal data",
			storeFunc: func(t *testing.T) SubRepoPermsStore {
				t.Helper()

				s := newStore(t)
				err := s.Upsert(context.Background(), 1, 1, authz.SubRepoPermissions{
					Paths: []string{
						"/not/the/same/*",
						"-/unlike/other/*",
					},
				})

				if err != nil {
					t.Fatal(err)
				}
				return s
			},
		},
		{
			name: "store with initial IP data",
			storeFunc: func(t *testing.T) SubRepoPermsStore {
				t.Helper()

				s := newStore(t)
				err := s.UpsertWithIPs(context.Background(), 1, 1, authz.SubRepoPermissionsWithIPs{
					Paths: []authz.PathWithIP{
						{Path: "/not/the/same/*", IP: "2001:db8::1"},
						{Path: "-/unlike/other/*", IP: "2001:db8::1"},
					},
				})

				if err != nil {
					t.Fatal(err)
				}
				return s
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			t.Run("UpsertWithIP", func(t *testing.T) {
				t.Run("Get (normal)", func(t *testing.T) {
					t.Parallel()

					inputPerms := authz.SubRepoPermissionsWithIPs{
						Paths: []authz.PathWithIP{
							{Path: "/src/foo/*", IP: "192.168.1.1"},
							{Path: "-/src/bar/*", IP: "192.168.1.2"},
						},
					}

					s := test.storeFunc(t)

					if err := s.UpsertWithIPs(ctx, userID, repoID, inputPerms); err != nil {
						t.Fatal(err)
					}

					actualPerms, err := s.Get(ctx, userID, repoID)
					if err != nil {
						t.Fatal(err)
					}

					var expectedPerms authz.SubRepoPermissions
					for _, p := range inputPerms.Paths {
						expectedPerms.Paths = append(expectedPerms.Paths, p.Path)
					}

					if diff := cmp.Diff(&expectedPerms, actualPerms); diff != "" {
						t.Fatalf("unexpected permissions (-want +got):\n%s", diff)
					}
				})

				t.Run("GetWithIP", func(t *testing.T) {
					for _, backfill := range []bool{true, false} {
						t.Run(fmt.Sprintf("backfill=%t", backfill), func(t *testing.T) {
							t.Parallel()

							inputPerms := authz.SubRepoPermissionsWithIPs{
								Paths: []authz.PathWithIP{
									{Path: "/src/foo/*", IP: "192.168.1.1"},
									{Path: "-/src/bar/*", IP: "192.168.1.2"},
								},
							}

							s := test.storeFunc(t)

							if err := s.UpsertWithIPs(ctx, userID, repoID, inputPerms); err != nil {
								t.Fatal(err)
							}

							actualPerms, err := s.GetWithIPs(ctx, userID, repoID, backfill)
							if err != nil {
								t.Fatal(err)
							}

							expectedPerms := inputPerms

							if diff := cmp.Diff(&expectedPerms, actualPerms); diff != "" {
								t.Fatalf("unexpected permissions (-want +got):\n%s", diff)
							}
						})
					}
				})

			})

			t.Run("Upsert (normal)", func(t *testing.T) {
				t.Run("Get (normal)", func(t *testing.T) {
					t.Parallel()

					inputPerms := authz.SubRepoPermissions{
						Paths: []string{
							"/src/foo/*",
							"-/src/bar/*",
						},
					}

					s := test.storeFunc(t)

					if err := s.Upsert(ctx, userID, repoID, inputPerms); err != nil {
						t.Fatal(err)
					}

					actualPerms, err := s.Get(ctx, userID, repoID)
					if err != nil {
						t.Fatal(err)
					}

					expectedPerms := inputPerms
					if diff := cmp.Diff(&expectedPerms, actualPerms); diff != "" {
						t.Fatalf("unexpected permissions (-want +got):\n%s", diff)
					}
				})

				t.Run("GetWithIP", func(t *testing.T) {

					t.Run("backfill=true", func(t *testing.T) {
						t.Parallel()

						inputPerms := authz.SubRepoPermissions{
							Paths: []string{
								"/src/foo/*",
								"-/src/bar/*",
							},
						}

						s := test.storeFunc(t)
						err := s.Upsert(ctx, userID, repoID, inputPerms)

						if err != nil {
							t.Fatal(err)
						}

						actualPerms, err := s.GetWithIPs(ctx, userID, repoID, true)
						if err != nil {
							t.Fatalf("unexpected error: %s", err)
						}

						expected := authz.SubRepoPermissionsWithIPs{
							Paths: []authz.PathWithIP{
								{Path: "/src/foo/*", IP: "*"},
								{Path: "-/src/bar/*", IP: "*"},
							},
						}

						if diff := cmp.Diff(&expected, actualPerms); diff != "" {
							t.Fatalf("unexpected permissions (-want +got):\n%s", diff)
						}
					})

					t.Run("backfill=false", func(t *testing.T) {
						t.Parallel()

						inputPerms := authz.SubRepoPermissions{
							Paths: []string{
								"/src/foo/*",
								"-/src/bar/*",
							},
						}

						s := test.storeFunc(t)
						err := s.Upsert(ctx, userID, repoID, inputPerms)

						if err != nil {
							t.Fatal(err)
						}

						_, err = s.GetWithIPs(ctx, userID, repoID, false)
						if !errors.Is(err, IPsNotSyncedError) {
							t.Fatalf("unexpected error: %s", err)
						}
					})

				})
			})
		})
	}
}

func TestSubRepoPermsStore_GetByUserWithIPs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	newStore := func(t *testing.T) (SubRepoPermsStore, DB) {
		t.Helper()

		logger := logtest.Scoped(t)
		db := NewDB(logger, dbtest.NewDB(t))

		ctx := context.Background()
		prepareSubRepoTestData(ctx, t, db)
		s := db.SubRepoPerms()

		return s, db
	}

	userID := int32(1)

	ctx := context.Background()

	for _, test := range []struct {
		name      string
		storeFunc func(t *testing.T) (SubRepoPermsStore, DB)
	}{
		{
			name:      "empty store",
			storeFunc: newStore,
		},
		{
			name: "store with initial normal data",
			storeFunc: func(t *testing.T) (SubRepoPermsStore, DB) {
				t.Helper()

				s, db := newStore(t)
				err := s.Upsert(context.Background(), 1, 1, authz.SubRepoPermissions{
					Paths: []string{
						"/not/the/same/*",
						"-/unlike/other/*",
					},
				})

				if err != nil {
					t.Fatal(err)
				}

				err = s.Upsert(context.Background(), 1, 2, authz.SubRepoPermissions{
					Paths: []string{
						"/not/the/same/*",
						"-/unlike/other/*",
					},
				})

				if err != nil {
					t.Fatal(err)
				}
				return s, db
			},
		},
		{
			name: "store with initial IP data",
			storeFunc: func(t *testing.T) (SubRepoPermsStore, DB) {
				t.Helper()

				s, db := newStore(t)
				err := s.UpsertWithIPs(context.Background(), 1, 1, authz.SubRepoPermissionsWithIPs{
					Paths: []authz.PathWithIP{
						{Path: "/not/the/same/*", IP: "2001:db8::1"},
						{Path: "-/unlike/other/*", IP: "2001:db8::1"},
					},
				})
				if err != nil {
					t.Fatal(err)
				}

				err = s.UpsertWithIPs(context.Background(), 1, 2, authz.SubRepoPermissionsWithIPs{
					Paths: []authz.PathWithIP{
						{Path: "/not/the/same/*", IP: "2001:db8::1"},
						{Path: "-/unlike/other/*", IP: "2001:db8::1"},
					},
				})

				if err != nil {
					t.Fatal(err)
				}
				return s, db
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Run("UpsertWithIP", func(t *testing.T) {

				t.Run("GetByUser (normal)", func(t *testing.T) {
					conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthzEnforceForSiteAdmins: true}})
					t.Cleanup(func() { conf.Mock(nil) })

					inputPermsMap := map[api.RepoID]struct {
						name  api.RepoName
						perms authz.SubRepoPermissionsWithIPs
					}{
						1: {
							name: "github.com/foo/bar",
							perms: authz.SubRepoPermissionsWithIPs{
								Paths: []authz.PathWithIP{
									{Path: "/src/foo/*", IP: "192.168.1.1"},
									{Path: "-/src/bar/*", IP: "192.168.1.2"},
								},
							},
						},

						2: {
							name: "github.com/foo/baz",
							perms: authz.SubRepoPermissionsWithIPs{
								Paths: []authz.PathWithIP{
									{Path: "/src/baz/*", IP: "192.167.1.1"},
									{Path: "-/src/quiz/*", IP: "192.167.1.2"},
								},
							},
						},
					}

					s, db := test.storeFunc(t)

					for repoID, inputPerms := range inputPermsMap {
						if err := s.UpsertWithIPs(ctx, userID, repoID, inputPerms.perms); err != nil {
							t.Fatalf("unexpected error when inserting permissions for repo %d: %s", repoID, err)
						}
					}

					actualPerms, err := s.GetByUser(ctx, userID)
					if err != nil {
						t.Fatal(err)
					}

					expectedPerms := make(map[api.RepoName]authz.SubRepoPermissions)

					for _, entry := range inputPermsMap {
						var perms authz.SubRepoPermissions
						for _, p := range entry.perms.Paths {
							perms.Paths = append(perms.Paths, p.Path)
						}
						expectedPerms[entry.name] = perms
					}

					if diff := cmp.Diff(expectedPerms, actualPerms); diff != "" {
						t.Fatalf("unexpected permissions (-want +got):\n%s", diff)
					}

					for _, tc := range []struct {
						siteAdmin           bool
						enforceForSiteAdmin bool
						wantRows            bool
					}{
						{siteAdmin: true, enforceForSiteAdmin: true, wantRows: true},
						{siteAdmin: false, enforceForSiteAdmin: false, wantRows: true},
						{siteAdmin: true, enforceForSiteAdmin: false, wantRows: false},
						{siteAdmin: false, enforceForSiteAdmin: true, wantRows: true},
					} {
						t.Run(fmt.Sprintf("SiteAdmin:%t-Enforce:%t", tc.siteAdmin, tc.enforceForSiteAdmin), func(t *testing.T) {
							conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthzEnforceForSiteAdmins: tc.enforceForSiteAdmin}})
							result, err := db.ExecContext(ctx, "UPDATE users SET site_admin = $1 WHERE id = $2", tc.siteAdmin, userID)
							if err != nil {
								t.Fatal(err)
							}

							affected, err := result.RowsAffected()
							if err != nil {
								t.Fatal(err)
							}
							if affected != 1 {
								t.Fatalf("Wanted 1 row affected, got %d", affected)
							}

							have, err := s.GetByUser(ctx, userID)
							if err != nil {
								t.Fatal(err)
							}
							if tc.wantRows {
								assert.NotEmpty(t, have)
							} else {
								assert.Empty(t, have)
							}
						})
					}
				})

				t.Run("GetByUserWithIP", func(t *testing.T) {
					for _, backfill := range []bool{true, false} {
						t.Run(fmt.Sprintf("backfill=%t", backfill), func(t *testing.T) {
							conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthzEnforceForSiteAdmins: true}})
							t.Cleanup(func() { conf.Mock(nil) })

							inputPermsMap := map[api.RepoID]struct {
								name  api.RepoName
								perms authz.SubRepoPermissionsWithIPs
							}{
								1: {
									name: "github.com/foo/bar",
									perms: authz.SubRepoPermissionsWithIPs{
										Paths: []authz.PathWithIP{
											{Path: "/src/foo/*", IP: "192.168.1.1"},
											{Path: "-/src/bar/*", IP: "192.168.1.2"},
										},
									},
								},

								2: {
									name: "github.com/foo/baz",
									perms: authz.SubRepoPermissionsWithIPs{
										Paths: []authz.PathWithIP{
											{Path: "/src/baz/*", IP: "192.167.1.1"},
											{Path: "-/src/quiz/*", IP: "192.167.1.2"},
										},
									},
								},
							}

							s, db := test.storeFunc(t)

							for repoID, inputPerms := range inputPermsMap {
								if err := s.UpsertWithIPs(ctx, userID, repoID, inputPerms.perms); err != nil {
									t.Fatalf("unexpected error when inserting permissions for repo %d: %s", repoID, err)
								}
							}

							expectedPermsWithIPs := make(map[api.RepoName]authz.SubRepoPermissionsWithIPs)
							for _, entry := range inputPermsMap {
								expectedPermsWithIPs[entry.name] = entry.perms
							}

							actualPermsWithIPs, err := s.GetByUserWithIPs(ctx, userID, backfill)
							if err != nil {
								t.Fatal(err)
							}

							if diff := cmp.Diff(expectedPermsWithIPs, actualPermsWithIPs); diff != "" {
								t.Fatalf("unexpected permissions (-want +got):\n%s", diff)
							}

							for _, tc := range []struct {
								siteAdmin           bool
								enforceForSiteAdmin bool
								wantRows            bool
							}{
								{siteAdmin: true, enforceForSiteAdmin: true, wantRows: true},
								{siteAdmin: false, enforceForSiteAdmin: false, wantRows: true},
								{siteAdmin: true, enforceForSiteAdmin: false, wantRows: false},
								{siteAdmin: false, enforceForSiteAdmin: true, wantRows: true},
							} {
								t.Run(fmt.Sprintf("SiteAdmin:%t-Enforce:%t", tc.siteAdmin, tc.enforceForSiteAdmin), func(t *testing.T) {
									conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthzEnforceForSiteAdmins: tc.enforceForSiteAdmin}})
									result, err := db.ExecContext(ctx, "UPDATE users SET site_admin = $1 WHERE id = $2", tc.siteAdmin, userID)
									if err != nil {
										t.Fatal(err)
									}

									affected, err := result.RowsAffected()
									if err != nil {
										t.Fatal(err)
									}
									if affected != 1 {
										t.Fatalf("Wanted 1 row affected, got %d", affected)
									}

									have, err := s.GetByUserWithIPs(ctx, userID, backfill)
									if err != nil {
										t.Fatal(err)
									}
									if tc.wantRows {
										assert.NotEmpty(t, have)
									} else {
										assert.Empty(t, have)
									}
								})
							}
						})
					}
				})
			})

			t.Run("Upsert (normal)", func(t *testing.T) {
				t.Run("GetByUser (normal)", func(t *testing.T) {
					conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthzEnforceForSiteAdmins: true}})
					t.Cleanup(func() { conf.Mock(nil) })

					inputMap := map[api.RepoID]struct {
						name  api.RepoName
						perms authz.SubRepoPermissions
					}{
						1: {
							name: "github.com/foo/bar",
							perms: authz.SubRepoPermissions{
								Paths: []string{
									"/src/foo/*",
									"-/src/bar/*",
								},
							},
						},

						2: {
							name: "github.com/foo/baz",
							perms: authz.SubRepoPermissions{
								Paths: []string{
									"/src/baz/*",
									"-/src/quiz/*",
								},
							},
						},
					}

					s, db := test.storeFunc(t)

					for repoID, input := range inputMap {
						if err := s.Upsert(ctx, userID, repoID, input.perms); err != nil {
							t.Fatalf("unexpected error when inserting permissions for repo %d: %s", repoID, err)
						}
					}

					actualPerms, err := s.GetByUser(ctx, userID)
					if err != nil {
						t.Fatal(err)
					}

					expectedPerms := make(map[api.RepoName]authz.SubRepoPermissions)
					for _, entry := range inputMap {
						var perms authz.SubRepoPermissions
						perms.Paths = append(perms.Paths, entry.perms.Paths...)
						expectedPerms[entry.name] = perms
					}

					if diff := cmp.Diff(expectedPerms, actualPerms); diff != "" {
						t.Fatalf("unexpected permissions (-want +got):\n%s", diff)
					}

					for _, tc := range []struct {
						siteAdmin           bool
						enforceForSiteAdmin bool
						wantRows            bool
					}{
						{siteAdmin: true, enforceForSiteAdmin: true, wantRows: true},
						{siteAdmin: false, enforceForSiteAdmin: false, wantRows: true},
						{siteAdmin: true, enforceForSiteAdmin: false, wantRows: false},
						{siteAdmin: false, enforceForSiteAdmin: true, wantRows: true},
					} {
						t.Run(fmt.Sprintf("SiteAdmin:%t-Enforce:%t", tc.siteAdmin, tc.enforceForSiteAdmin), func(t *testing.T) {
							conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthzEnforceForSiteAdmins: tc.enforceForSiteAdmin}})
							result, err := db.ExecContext(ctx, "UPDATE users SET site_admin = $1 WHERE id = $2", tc.siteAdmin, userID)
							if err != nil {
								t.Fatal(err)
							}

							affected, err := result.RowsAffected()
							if err != nil {
								t.Fatal(err)
							}
							if affected != 1 {
								t.Fatalf("Wanted 1 row affected, got %d", affected)
							}

							have, err := s.GetByUser(ctx, userID)
							if err != nil {
								t.Fatal(err)
							}
							if tc.wantRows {
								assert.NotEmpty(t, have)
							} else {
								assert.Empty(t, have)
							}
						})
					}
				})

				t.Run("GetByUserWithIP", func(t *testing.T) {
					t.Run("backfill=true", func(t *testing.T) {
						conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthzEnforceForSiteAdmins: true}})
						t.Cleanup(func() { conf.Mock(nil) })

						inputMap := map[api.RepoID]struct {
							name  api.RepoName
							perms authz.SubRepoPermissions
						}{
							1: {
								name: "github.com/foo/bar",
								perms: authz.SubRepoPermissions{
									Paths: []string{
										"/src/foo/*",
										"-/src/bar/*",
									},
								},
							},

							2: {
								name: "github.com/foo/baz",
								perms: authz.SubRepoPermissions{
									Paths: []string{
										"/src/baz/*",
										"-/src/quiz/*",
									},
								},
							},
						}

						s, db := test.storeFunc(t)

						for repoID, input := range inputMap {
							if err := s.Upsert(ctx, userID, repoID, input.perms); err != nil {
								t.Fatalf("unexpected error when inserting permissions for repo %d: %s", repoID, err)
							}
						}

						expectedPermsWithIPs := make(map[api.RepoName]authz.SubRepoPermissionsWithIPs)
						for _, entry := range inputMap {
							perms := authz.SubRepoPermissionsWithIPs{}

							for _, p := range entry.perms.Paths {
								perms.Paths = append(perms.Paths, authz.PathWithIP{Path: p, IP: "*"})
							}

							expectedPermsWithIPs[entry.name] = perms
						}

						actualPermsWithIPs, err := s.GetByUserWithIPs(ctx, userID, true)
						if err != nil {
							t.Fatal(err)
						}

						if diff := cmp.Diff(expectedPermsWithIPs, actualPermsWithIPs); diff != "" {
							t.Fatalf("unexpected permissions (-want +got):\n%s", diff)
						}

						for _, tc := range []struct {
							siteAdmin           bool
							enforceForSiteAdmin bool
							wantRows            bool
						}{
							{siteAdmin: true, enforceForSiteAdmin: true, wantRows: true},
							{siteAdmin: false, enforceForSiteAdmin: false, wantRows: true},
							{siteAdmin: true, enforceForSiteAdmin: false, wantRows: false},
							{siteAdmin: false, enforceForSiteAdmin: true, wantRows: true},
						} {
							t.Run(fmt.Sprintf("SiteAdmin:%t-Enforce:%t", tc.siteAdmin, tc.enforceForSiteAdmin), func(t *testing.T) {
								conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthzEnforceForSiteAdmins: tc.enforceForSiteAdmin}})
								result, err := db.ExecContext(ctx, "UPDATE users SET site_admin = $1 WHERE id = $2", tc.siteAdmin, userID)
								if err != nil {
									t.Fatal(err)
								}

								affected, err := result.RowsAffected()
								if err != nil {
									t.Fatal(err)
								}
								if affected != 1 {
									t.Fatalf("Wanted 1 row affected, got %d", affected)
								}

								have, err := s.GetByUserWithIPs(ctx, userID, true)
								if err != nil {
									t.Fatal(err)
								}
								if tc.wantRows {
									assert.NotEmpty(t, have)
								} else {
									assert.Empty(t, have)
								}
							})
						}
					})

					t.Run("backfill=false", func(t *testing.T) {
						conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthzEnforceForSiteAdmins: true}})
						t.Cleanup(func() { conf.Mock(nil) })

						inputMap := map[api.RepoID]struct {
							name  api.RepoName
							perms authz.SubRepoPermissions
						}{
							1: {
								name: "github.com/foo/bar",
								perms: authz.SubRepoPermissions{
									Paths: []string{
										"/src/foo/*",
										"-/src/bar/*",
									},
								},
							},

							2: {
								name: "github.com/foo/baz",
								perms: authz.SubRepoPermissions{
									Paths: []string{
										"/src/baz/*",
										"-/src/quiz/*",
									},
								},
							},
						}

						s, db := test.storeFunc(t)

						for repoID, input := range inputMap {
							if err := s.Upsert(ctx, userID, repoID, input.perms); err != nil {
								t.Fatalf("unexpected error when inserting permissions for repo %d: %s", repoID, err)
							}
						}

						_, err := s.GetByUserWithIPs(ctx, userID, false)
						if !errors.Is(err, IPsNotSyncedError) {
							t.Fatalf("unexpected error: %s", err)
						}

						for _, tc := range []struct {
							siteAdmin           bool
							enforceForSiteAdmin bool
							wantRows            bool
							wantErr             error
						}{
							{siteAdmin: true, enforceForSiteAdmin: true, wantErr: IPsNotSyncedError},
							{siteAdmin: false, enforceForSiteAdmin: false, wantErr: IPsNotSyncedError},

							// In the case where enforceForSiteAdmin is false, we want to return an
							// empty result set instead of an error so that we can force a fallback
							// to repo-level checks (so that all paths are accessible).
							{siteAdmin: true, enforceForSiteAdmin: false, wantRows: false, wantErr: nil},

							{siteAdmin: false, enforceForSiteAdmin: true, wantErr: IPsNotSyncedError},
						} {
							t.Run(fmt.Sprintf("SiteAdmin:%t-Enforce:%t", tc.siteAdmin, tc.enforceForSiteAdmin), func(t *testing.T) {
								conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthzEnforceForSiteAdmins: tc.enforceForSiteAdmin}})
								result, err := db.ExecContext(ctx, "UPDATE users SET site_admin = $1 WHERE id = $2", tc.siteAdmin, userID)
								if err != nil {
									t.Fatal(err)
								}

								affected, err := result.RowsAffected()
								if err != nil {
									t.Fatal(err)
								}
								if affected != 1 {
									t.Fatalf("Wanted 1 row affected, got %d", affected)
								}

								have, err := s.GetByUserWithIPs(ctx, userID, false)
								if tc.wantErr != nil {
									assert.ErrorIs(t, err, tc.wantErr)
									return
								}

								assert.NoError(t, err)
								if tc.wantRows {
									assert.NotEmpty(t, have)
								} else {
									assert.Empty(t, have)
								}
							})
						}
					})
				})
			})
		})
	}
}

func TestSubRepoPerms_GetByUserAndServiceWithIP(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	newStore := func(t *testing.T) SubRepoPermsStore {
		t.Helper()

		logger := logtest.Scoped(t)
		db := NewDB(logger, dbtest.NewDB(t))

		ctx := context.Background()
		prepareSubRepoTestData(ctx, t, db)
		s := db.SubRepoPerms()

		return s
	}

	userID := int32(1)
	ctx := context.Background()

	for _, test := range []struct {
		name      string
		storeFunc func(t *testing.T) SubRepoPermsStore
	}{
		{
			name:      "empty store",
			storeFunc: newStore,
		},
		{
			name: "store with initial normal data",
			storeFunc: func(t *testing.T) SubRepoPermsStore {
				t.Helper()

				s := newStore(t)
				err := s.Upsert(context.Background(), 1, 1, authz.SubRepoPermissions{
					Paths: []string{
						"/not/the/same/*",
						"-/unlike/other/*",
					},
				})
				if err != nil {
					t.Fatal(err)
				}

				err = s.Upsert(context.Background(), 1, 2, authz.SubRepoPermissions{
					Paths: []string{
						"/not/the/same/*",
						"-/unlike/other/*",
					},
				})

				if err != nil {
					t.Fatal(err)
				}
				return s
			},
		},
		{
			name: "store with initial IP data",
			storeFunc: func(t *testing.T) SubRepoPermsStore {
				t.Helper()

				s := newStore(t)
				err := s.UpsertWithIPs(context.Background(), 1, 1, authz.SubRepoPermissionsWithIPs{
					Paths: []authz.PathWithIP{
						{Path: "/not/the/same/*", IP: "2001:db8::1"},
						{Path: "-/unlike/other/*", IP: "2001:db8::1"},
					},
				})
				if err != nil {
					t.Fatal(err)
				}

				err = s.UpsertWithIPs(context.Background(), 1, 2, authz.SubRepoPermissionsWithIPs{
					Paths: []authz.PathWithIP{
						{Path: "/not/the/same/*", IP: "2001:db8::1"},
						{Path: "-/unlike/other/*", IP: "2001:db8::1"},
					},
				})

				if err != nil {
					t.Fatal(err)
				}
				return s
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Run("Upsert (normal)", func(t *testing.T) {
				t.Run("GetByUserAndService (normal)", func(t *testing.T) {
					s := test.storeFunc(t)

					perms := authz.SubRepoPermissions{
						Paths: []string{"/src/foo/*", "-/src/bar/*"},
					}
					if err := s.Upsert(ctx, userID, api.RepoID(1), perms); err != nil {
						t.Fatal(err)
					}

					userID = int32(1)
					perms = authz.SubRepoPermissions{
						Paths: []string{"/src/foo2/*", "-/src/bar2/*"},
					}
					if err := s.Upsert(ctx, userID, api.RepoID(2), perms); err != nil {
						t.Fatal(err)
					}

					for _, tc := range []struct {
						name        string
						userID      int32
						serviceType string
						serviceID   string
						want        map[api.ExternalRepoSpec]authz.SubRepoPermissions
					}{
						{
							name:        "Unknown service",
							userID:      userID,
							serviceType: "abc",
							serviceID:   "xyz",
							want:        map[api.ExternalRepoSpec]authz.SubRepoPermissions{},
						},
						{
							name:        "Known service",
							userID:      userID,
							serviceType: "github",
							serviceID:   "https://github.com/",
							want: map[api.ExternalRepoSpec]authz.SubRepoPermissions{
								{
									ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
									ServiceType: "github",
									ServiceID:   "https://github.com/",
								}: {
									Paths: []string{"/src/foo/*", "-/src/bar/*"},
								},
								{
									ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOB==",
									ServiceType: "github",
									ServiceID:   "https://github.com/",
								}: {
									Paths: []string{"/src/foo2/*", "-/src/bar2/*"},
								},
							},
						},
					} {
						t.Run(tc.name, func(t *testing.T) {
							have, err := s.GetByUserAndService(ctx, userID, tc.serviceType, tc.serviceID)
							if err != nil {
								t.Fatal(err)
							}
							if diff := cmp.Diff(tc.want, have); diff != "" {
								t.Fatal(diff)
							}
						})
					}

				})

				t.Run("GetByUserAndServiceWithIP", func(t *testing.T) {
					t.Run("backfill=true", func(t *testing.T) {
						s := test.storeFunc(t)

						perms := authz.SubRepoPermissions{
							Paths: []string{"/src/foo/*", "-/src/bar/*"},
						}
						if err := s.Upsert(ctx, userID, api.RepoID(1), perms); err != nil {
							t.Fatal(err)
						}

						userID = int32(1)
						perms = authz.SubRepoPermissions{
							Paths: []string{"/src/foo2/*", "-/src/bar2/*"},
						}
						if err := s.Upsert(ctx, userID, api.RepoID(2), perms); err != nil {
							t.Fatal(err)
						}

						for _, tc := range []struct {
							name        string
							userID      int32
							serviceType string
							serviceID   string
							want        map[api.ExternalRepoSpec]authz.SubRepoPermissionsWithIPs
						}{
							{
								name:        "Unknown service",
								userID:      userID,
								serviceType: "abc",
								serviceID:   "xyz",
								want:        map[api.ExternalRepoSpec]authz.SubRepoPermissionsWithIPs{},
							},
							{
								name:        "Known service",
								userID:      userID,
								serviceType: "github",
								serviceID:   "https://github.com/",
								want: map[api.ExternalRepoSpec]authz.SubRepoPermissionsWithIPs{
									{
										ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
										ServiceType: "github",
										ServiceID:   "https://github.com/",
									}: {
										Paths: []authz.PathWithIP{
											{Path: "/src/foo/*", IP: "*"},
											{Path: "-/src/bar/*", IP: "*"},
										},
									},
									{
										ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOB==",
										ServiceType: "github",
										ServiceID:   "https://github.com/",
									}: {
										Paths: []authz.PathWithIP{
											{Path: "/src/foo2/*", IP: "*"},
											{Path: "-/src/bar2/*", IP: "*"},
										},
									},
								},
							},
						} {
							t.Run(tc.name, func(t *testing.T) {
								have, err := s.GetByUserAndServiceWithIPs(ctx, userID, tc.serviceType, tc.serviceID, true)

								assert.NoError(t, err)
								if diff := cmp.Diff(tc.want, have); diff != "" {
									t.Fatal(diff)
								}
							})
						}
					})

					t.Run("backfill=false", func(t *testing.T) {
						s := test.storeFunc(t)

						perms := authz.SubRepoPermissions{
							Paths: []string{"/src/foo/*", "-/src/bar/*"},
						}
						if err := s.Upsert(ctx, userID, api.RepoID(1), perms); err != nil {
							t.Fatal(err)
						}

						userID = int32(1)
						perms = authz.SubRepoPermissions{
							Paths: []string{"/src/foo2/*", "-/src/bar2/*"},
						}
						if err := s.Upsert(ctx, userID, api.RepoID(2), perms); err != nil {
							t.Fatal(err)
						}

						for _, tc := range []struct {
							name        string
							userID      int32
							serviceType string
							serviceID   string
							want        map[api.ExternalRepoSpec]authz.SubRepoPermissionsWithIPs
							wantErr     error
						}{
							{
								name:        "Unknown service",
								userID:      userID,
								serviceType: "abc",
								serviceID:   "xyz",
								want:        map[api.ExternalRepoSpec]authz.SubRepoPermissionsWithIPs{},
							},
							{
								name:        "Known service",
								userID:      userID,
								serviceType: "github",
								serviceID:   "https://github.com/",
								wantErr:     IPsNotSyncedError,
							},
						} {
							t.Run(tc.name, func(t *testing.T) {
								have, err := s.GetByUserAndServiceWithIPs(ctx, userID, tc.serviceType, tc.serviceID, false)
								if tc.wantErr != nil {
									assert.ErrorIs(t, err, tc.wantErr)
									return
								}

								assert.NoError(t, err)
								if diff := cmp.Diff(tc.want, have); diff != "" {
									t.Fatal(diff)
								}
							})
						}
					})

				})
			})

			t.Run("UpsertWithIP", func(t *testing.T) {
				t.Run("GetByUserAndService (normal)", func(t *testing.T) {
					s := test.storeFunc(t)

					perms := authz.SubRepoPermissionsWithIPs{
						Paths: []authz.PathWithIP{
							{Path: "/src/foo/*", IP: "192.168.1.1"},
							{Path: "-/src/bar/*", IP: "192.168.1.2"},
						},
					}
					if err := s.UpsertWithIPs(ctx, userID, api.RepoID(1), perms); err != nil {
						t.Fatal(err)
					}

					userID = int32(1)
					perms = authz.SubRepoPermissionsWithIPs{
						Paths: []authz.PathWithIP{
							{Path: "/src/foo2/*", IP: "10.0.0.1"},
							{Path: "-/src/bar2/*", IP: "10.0.0.2"},
						},
					}

					if err := s.UpsertWithIPs(ctx, userID, api.RepoID(2), perms); err != nil {
						t.Fatal(err)
					}

					for _, tc := range []struct {
						name        string
						userID      int32
						serviceType string
						serviceID   string
						want        map[api.ExternalRepoSpec]authz.SubRepoPermissions
					}{
						{
							name:        "Unknown service",
							userID:      userID,
							serviceType: "abc",
							serviceID:   "xyz",
							want:        map[api.ExternalRepoSpec]authz.SubRepoPermissions{},
						},
						{
							name:        "Known service",
							userID:      userID,
							serviceType: "github",
							serviceID:   "https://github.com/",
							want: map[api.ExternalRepoSpec]authz.SubRepoPermissions{
								{
									ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
									ServiceType: "github",
									ServiceID:   "https://github.com/",
								}: {
									Paths: []string{"/src/foo/*", "-/src/bar/*"},
								},
								{
									ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOB==",
									ServiceType: "github",
									ServiceID:   "https://github.com/",
								}: {
									Paths: []string{"/src/foo2/*", "-/src/bar2/*"},
								},
							},
						},
					} {
						t.Run(tc.name, func(t *testing.T) {
							have, err := s.GetByUserAndService(ctx, userID, tc.serviceType, tc.serviceID)
							if err != nil {
								t.Fatal(err)
							}
							if diff := cmp.Diff(tc.want, have); diff != "" {
								t.Fatal(diff)
							}
						})
					}
				})

				t.Run("GetByUserAndServiceWithIPs", func(t *testing.T) {
					for _, backfill := range []bool{true, false} {
						t.Run(fmt.Sprintf("backfill=%t", backfill), func(t *testing.T) {
							s := test.storeFunc(t)

							perms := authz.SubRepoPermissionsWithIPs{
								Paths: []authz.PathWithIP{
									{Path: "/src/foo/*", IP: "192.168.1.1"},
									{Path: "-/src/bar/*", IP: "192.168.1.2"},
								},
							}
							if err := s.UpsertWithIPs(ctx, userID, api.RepoID(1), perms); err != nil {
								t.Fatal(err)
							}

							userID = int32(1)
							perms = authz.SubRepoPermissionsWithIPs{
								Paths: []authz.PathWithIP{
									{Path: "/src/foo2/*", IP: "10.0.0.1"},
									{Path: "-/src/bar2/*", IP: "10.0.0.2"},
								},
							}

							if err := s.UpsertWithIPs(ctx, userID, api.RepoID(2), perms); err != nil {
								t.Fatal(err)
							}

							for _, tc := range []struct {
								name        string
								userID      int32
								serviceType string
								serviceID   string
								want        map[api.ExternalRepoSpec]authz.SubRepoPermissionsWithIPs
							}{
								{
									name:        "Unknown service",
									userID:      userID,
									serviceType: "abc",
									serviceID:   "xyz",
									want:        map[api.ExternalRepoSpec]authz.SubRepoPermissionsWithIPs{},
								},
								{
									name:        "Known service",
									userID:      userID,
									serviceType: "github",
									serviceID:   "https://github.com/",
									want: map[api.ExternalRepoSpec]authz.SubRepoPermissionsWithIPs{
										{
											ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
											ServiceType: "github",
											ServiceID:   "https://github.com/",
										}: {
											Paths: []authz.PathWithIP{
												{Path: "/src/foo/*", IP: "192.168.1.1"},
												{Path: "-/src/bar/*", IP: "192.168.1.2"},
											},
										},
										{
											ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOB==",
											ServiceType: "github",
											ServiceID:   "https://github.com/",
										}: {
											Paths: []authz.PathWithIP{
												{Path: "/src/foo2/*", IP: "10.0.0.1"},
												{Path: "-/src/bar2/*", IP: "10.0.0.2"},
											},
										},
									},
								},
							} {
								t.Run(tc.name, func(t *testing.T) {
									have, err := s.GetByUserAndServiceWithIPs(ctx, userID, tc.serviceType, tc.serviceID, backfill)
									if err != nil {
										t.Fatal(err)
									}
									if diff := cmp.Diff(tc.want, have); diff != "" {
										t.Fatal(diff)
									}
								})
							}
						})
					}
				})
			})
		})
	}
}

func TestSubRepoPerms_UpsertWithSpecWithIP(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()
	userID := int32(1)
	repoID := api.RepoID(1)

	newStore := func(t *testing.T) SubRepoPermsStore {
		t.Helper()
		logger := logtest.Scoped(t)
		db := NewDB(logger, dbtest.NewDB(t))

		ctx := context.Background()
		prepareSubRepoTestData(ctx, t, db)
		s := db.SubRepoPerms()
		return s
	}

	for _, test := range []struct {
		name      string
		storeFunc func(t *testing.T) SubRepoPermsStore
	}{
		{
			name:      "empty store",
			storeFunc: newStore,
		},
		{
			name: "store with initial normal data",
			storeFunc: func(t *testing.T) SubRepoPermsStore {
				t.Helper()

				s := newStore(t)
				err := s.Upsert(context.Background(), 1, 1, authz.SubRepoPermissions{
					Paths: []string{
						"/not/the/same/*",
						"-/unlike/other/*",
					},
				})
				if err != nil {
					t.Fatal(err)
				}

				err = s.Upsert(context.Background(), 1, 2, authz.SubRepoPermissions{
					Paths: []string{
						"/not/the/same/*",
						"-/unlike/other/*",
					},
				})

				if err != nil {
					t.Fatal(err)
				}
				return s
			},
		},
		{
			name: "store with initial IP data",
			storeFunc: func(t *testing.T) SubRepoPermsStore {
				t.Helper()

				s := newStore(t)
				err := s.UpsertWithIPs(context.Background(), 1, 1, authz.SubRepoPermissionsWithIPs{
					Paths: []authz.PathWithIP{
						{Path: "/not/the/same/*", IP: "2001:db8::1"},
						{Path: "-/unlike/other/*", IP: "2001:db8::1"},
					},
				})

				if err != nil {
					t.Fatal(err)
				}

				err = s.UpsertWithIPs(context.Background(), 1, 2, authz.SubRepoPermissionsWithIPs{
					Paths: []authz.PathWithIP{
						{Path: "/not/the/same/*", IP: "2001:db8::1"},
						{Path: "-/unlike/other/*", IP: "2001:db8::1"},
					},
				})

				if err != nil {
					t.Fatal(err)
				}
				return s
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Run("UpsertWithSpec (normal)", func(t *testing.T) {
				t.Run("Get (normal)", func(t *testing.T) {
					for _, backfill := range []bool{true, false} {
						t.Run(fmt.Sprintf("backfill=%t", backfill), func(t *testing.T) {
							t.Parallel()

							s := test.storeFunc(t)

							perms := authz.SubRepoPermissions{
								Paths: []string{"/src/foo/*", "-/src/bar/*"},
							}
							err := s.UpsertWithSpec(ctx, userID, api.ExternalRepoSpec{
								ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
								ServiceType: "github",
								ServiceID:   "https://github.com/",
							}, perms)
							if err != nil {
								t.Fatal(err)
							}

							have, err := s.Get(ctx, userID, repoID)
							if err != nil {
								t.Fatal(err)
							}
							if diff := cmp.Diff(&perms, have); diff != "" {
								t.Fatal(diff)
							}
						})
					}
				})

				t.Run("GetWithIP", func(t *testing.T) {
					t.Run("backfill=true", func(t *testing.T) {
						t.Parallel()

						s := test.storeFunc(t)

						perms := authz.SubRepoPermissions{
							Paths: []string{"/src/foo/*", "-/src/bar/*"},
						}
						err := s.UpsertWithSpec(ctx, userID, api.ExternalRepoSpec{
							ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
							ServiceType: "github",
							ServiceID:   "https://github.com/",
						}, perms)
						if err != nil {
							t.Fatal(err)
						}

						expectedPerms := &authz.SubRepoPermissionsWithIPs{
							Paths: []authz.PathWithIP{
								{Path: "/src/foo/*", IP: "*"},
								{Path: "-/src/bar/*", IP: "*"},
							},
						}

						have, err := s.GetWithIPs(ctx, userID, repoID, true)
						if err != nil {
							t.Fatal(err)
						}

						if diff := cmp.Diff(expectedPerms, have); diff != "" {
							t.Fatalf("unexpected permissions (-want +got):\n%s", diff)
						}
					})

					t.Run("backfill=false", func(t *testing.T) {
						t.Parallel()

						s := test.storeFunc(t)

						perms := authz.SubRepoPermissions{
							Paths: []string{"/src/foo/*", "-/src/bar/*"},
						}
						err := s.UpsertWithSpec(ctx, userID, api.ExternalRepoSpec{
							ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
							ServiceType: "github",
							ServiceID:   "https://github.com/",
						}, perms)
						if err != nil {
							t.Fatal(err)
						}

						_, err = s.GetWithIPs(ctx, userID, repoID, false)
						if !errors.Is(err, IPsNotSyncedError) {
							t.Fatal(err)
						}
					})
				})
			})

			t.Run("UpsertWithSpecWithIPs", func(t *testing.T) {
				t.Run("Get (normal)", func(t *testing.T) {
					t.Parallel()

					s := test.storeFunc(t)

					perms := authz.SubRepoPermissionsWithIPs{
						Paths: []authz.PathWithIP{
							{Path: "/src/foo/*", IP: "192.168.1.1"},
							{Path: "-/src/bar/*", IP: "192.168.1.2"},
						},
					}

					err := s.UpsertWithSpecWithIPs(ctx, userID, api.ExternalRepoSpec{
						ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
						ServiceType: "github",
						ServiceID:   "https://github.com/",
					}, perms)
					if err != nil {
						t.Fatal(err)
					}

					expectedPerms := &authz.SubRepoPermissions{
						Paths: []string{"/src/foo/*", "-/src/bar/*"},
					}

					have, err := s.Get(ctx, userID, repoID)
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(expectedPerms, have); diff != "" {
						t.Fatalf("unexpected permissions (-want +got):\n%s", diff)
					}
				})

				t.Run("GetWithIP", func(t *testing.T) {
					for _, backfill := range []bool{true, false} {
						t.Run(fmt.Sprintf("backfill=%t", backfill), func(t *testing.T) {
							t.Parallel()

							s := test.storeFunc(t)

							perms := authz.SubRepoPermissionsWithIPs{
								Paths: []authz.PathWithIP{
									{Path: "/src/foo/*", IP: "192.168.1.1"},
									{Path: "-/src/bar/*", IP: "192.168.1.2"},
								},
							}

							err := s.UpsertWithSpecWithIPs(ctx, userID, api.ExternalRepoSpec{
								ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
								ServiceType: "github",
								ServiceID:   "https://github.com/",
							}, perms)
							if err != nil {
								t.Fatal(err)
							}

							have, err := s.GetWithIPs(ctx, userID, repoID, backfill)
							if err != nil {
								t.Fatal(err)
							}

							if diff := cmp.Diff(&perms, have); diff != "" {
								t.Fatalf("unexpected permissions (-want +got):\n%s", diff)
							}
						})
					}
				})
			})
		})
	}
}
