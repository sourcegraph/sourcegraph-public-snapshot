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

	type update struct {
		bindIDs []string
		repoID  int32
	}
	tests := []struct {
		name          string
		config        *schema.PermissionsUserMapping
		args          *db.GrantPendingPermissionsArgs
		updates       []update
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
			updates: []update{
				{
					bindIDs: []string{"alice@example.com"},
					repoID:  1,
				},
				{
					bindIDs: []string{"alice2@example.com"},
					repoID:  2,
				},
				{
					bindIDs: []string{"alice3@example.com"},
					repoID:  3,
				},
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
			updates: []update{
				{
					bindIDs: []string{"alice"},
					repoID:  1,
				},
				{
					bindIDs: []string{"bob"},
					repoID:  2,
				},
			},
			expectRepoIDs: []uint32{1},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer cleanupPermsTables(t, s.store)

			globals.SetPermissionsUserMapping(test.config)

			for _, update := range test.updates {
				err := s.store.SetRepoPendingPermissions(ctx, update.bindIDs, &iauthz.RepoPermissions{
					RepoID:   update.repoID,
					Perm:     authz.Read,
					Provider: authz.ProviderSourcegraph,
				})
				if err != nil {
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

	type update struct {
		repoID  int32
		userIDs []uint32
	}
	tests := []struct {
		name        string
		args        *db.AuthorizedReposArgs
		updates     []update
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
			updates: []update{
				{
					repoID:  1,
					userIDs: []uint32{1},
				},
				{
					repoID:  2,
					userIDs: []uint32{1},
				},
				{
					repoID:  3,
					userIDs: []uint32{1},
				},
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
			updates: []update{
				{
					repoID:  1,
					userIDs: []uint32{1},
				},
			},
			expectRepos: []*types.Repo{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer cleanupPermsTables(t, s.store)

			for _, update := range test.updates {
				err := s.store.SetRepoPermissions(ctx, &iauthz.RepoPermissions{
					RepoID:   update.repoID,
					Perm:     authz.Read,
					UserIDs:  toBitmap(update.userIDs...),
					Provider: authz.ProviderSourcegraph,
				})
				if err != nil {
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
