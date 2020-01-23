package db

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAuthzStore_GrantPendingPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	// Create user with initially verified email
	user, err := db.Users.Create(ctx, db.NewUser{
		Email:           "alice@example.com",
		Username:        "alice",
		EmailIsVerified: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	code := "verify-code"

	// Add and verify the second email
	err = db.UserEmails.Add(ctx, user.ID, "alice2@example.com", &code)
	if err != nil {
		t.Fatal(err)
	}
	err = db.UserEmails.SetVerified(ctx, user.ID, "alice2@example.com", true)
	if err != nil {
		t.Fatal(err)
	}

	// Add third email and leave as unverified
	err = db.UserEmails.Add(ctx, user.ID, "alice3@example.com", &code)
	if err != nil {
		t.Fatal(err)
	}

	s := NewAuthzStore(dbconn.Global, clock).(*authzStore)

	tests := []struct {
		name          string
		config        *schema.PermissionsUserMapping
		args          *db.GrantPendingPermissionsArgs
		update        func(ctx context.Context, s *PermsStore) error
		expectRepoIDs []uint32
	}{
		{
			name: "grant by emails",
			config: &schema.PermissionsUserMapping{
				BindID: "email",
			},
			args: &db.GrantPendingPermissionsArgs{
				UserID: user.ID,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			},
			update: func(ctx context.Context, s *PermsStore) error {
				err := s.SetRepoPendingPermissions(ctx, []string{"alice@example.com"}, &iauthz.RepoPermissions{
					RepoID:   1,
					Perm:     authz.Read,
					UserIDs:  toBitmap(1),
					Provider: authz.ProviderSourcegraph,
				})
				if err != nil {
					return err
				}

				err = s.SetRepoPendingPermissions(ctx, []string{"alice2@example.com"}, &iauthz.RepoPermissions{
					RepoID:   2,
					Perm:     authz.Read,
					UserIDs:  toBitmap(1),
					Provider: authz.ProviderSourcegraph,
				})
				if err != nil {
					return err
				}

				err = s.SetRepoPendingPermissions(ctx, []string{"alice3@example.com"}, &iauthz.RepoPermissions{
					RepoID:   3,
					Perm:     authz.Read,
					UserIDs:  toBitmap(1),
					Provider: authz.ProviderSourcegraph,
				})
				if err != nil {
					return err
				}
				return nil
			},
			expectRepoIDs: []uint32{1, 2},
		},
		{
			name: "grant by username",
			config: &schema.PermissionsUserMapping{
				BindID: "username",
			},
			args: &db.GrantPendingPermissionsArgs{
				UserID: user.ID,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			},
			update: func(ctx context.Context, s *PermsStore) error {
				err := s.SetRepoPendingPermissions(ctx, []string{"alice"}, &iauthz.RepoPermissions{
					RepoID:   1,
					Perm:     authz.Read,
					Provider: authz.ProviderSourcegraph,
				})
				if err != nil {
					return err
				}

				err = s.SetRepoPendingPermissions(ctx, []string{"bob"}, &iauthz.RepoPermissions{
					RepoID:   2,
					Perm:     authz.Read,
					Provider: authz.ProviderSourcegraph,
				})
				if err != nil {
					return err
				}

				return nil
			},
			expectRepoIDs: []uint32{1},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer cleanupPermsTables(t, s.store)

			globals.SetPermissionsUserMapping(test.config)

			if test.update != nil {
				if err = test.update(ctx, s.store); err != nil {
					t.Fatal(err)
				}
			}

			err := s.GrantPendingPermissions(ctx, test.args)
			if err != nil {
				t.Fatal(err)
			}

			p := &iauthz.UserPermissions{
				UserID:   user.ID,
				Perm:     authz.Read,
				Type:     authz.PermRepos,
				Provider: authz.ProviderSourcegraph,
			}
			err = s.store.LoadUserPermissions(ctx, p)
			if err != nil {
				t.Fatal(err)
			}

			equal(t, "p.IDs", test.expectRepoIDs, bitmapToArray(p.IDs))
		})
	}
}

func TestAuthzStore_AuthorizedRepos(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	s := NewAuthzStore(dbconn.Global, clock).(*authzStore)

	tests := []struct {
		name        string
		args        *db.AuthorizedReposArgs
		update      func(ctx context.Context, s *PermsStore) error
		expectRepos []*types.Repo
	}{
		{
			name: "no repos",
			args: &db.AuthorizedReposArgs{},
		},
		{
			name: "has permissions for user=1",
			args: &db.AuthorizedReposArgs{
				Repos: []*types.Repo{
					{ID: 1},
					{ID: 2},
					{ID: 4},
				},
				UserID:   1,
				Perm:     authz.Read,
				Type:     authz.PermRepos,
				Provider: authz.ProviderSourcegraph,
			},
			update: func(ctx context.Context, s *PermsStore) error {
				for _, repoID := range []int32{1, 2, 3} {
					err := s.SetRepoPermissions(ctx, &iauthz.RepoPermissions{
						RepoID:   repoID,
						Perm:     authz.Read,
						UserIDs:  toBitmap(1),
						Provider: authz.ProviderSourcegraph,
					})
					if err != nil {
						return err
					}
				}

				return nil
			},
			expectRepos: []*types.Repo{
				{ID: 1},
				{ID: 2},
			},
		},
		{
			name: "no permissions for user=2",
			args: &db.AuthorizedReposArgs{
				Repos: []*types.Repo{
					{ID: 1},
					{ID: 2},
				},
				UserID:   2,
				Perm:     authz.Read,
				Type:     authz.PermRepos,
				Provider: authz.ProviderSourcegraph,
			},
			update: func(ctx context.Context, s *PermsStore) error {
				for _, repoID := range []int32{1, 2, 3} {
					err := s.SetRepoPermissions(ctx, &iauthz.RepoPermissions{
						RepoID:   repoID,
						Perm:     authz.Read,
						UserIDs:  toBitmap(1),
						Provider: authz.ProviderSourcegraph,
					})
					if err != nil {
						return err
					}
				}

				return nil
			},
			expectRepos: []*types.Repo{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer cleanupPermsTables(t, s.store)

			if test.update != nil {
				if err := test.update(ctx, s.store); err != nil {
					t.Fatal(err)
				}
			}

			repos, err := s.AuthorizedRepos(ctx, test.args)
			if err != nil {
				t.Fatal(err)
			}

			equal(t, "repos", test.expectRepos, repos)
		})
	}
}
