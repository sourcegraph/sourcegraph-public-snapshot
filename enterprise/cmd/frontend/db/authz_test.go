package db

import (
	"context"
	"github.com/pkg/errors"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAuthzStore_GrantPendingPermissions(t *testing.T) {
	ctx := context.Background()
	s := NewAuthzStore(nil, clock)
	Mocks.Perms.Txs = func(context.Context) (*PermsStore, error) {
		return &PermsStore{}, nil
	}
	defer func() {
		Mocks.Perms.Txs = nil
	}()

	tests := []struct {
		name              string
		config            *schema.PermissionsUserMapping
		args              *db.GrantPendingPermissionsArgs
		mockEmails        []*db.UserEmail
		mockUser          *types.User
		expectCalledCount int
	}{
		{
			name:              "bad userID",
			config:            &schema.PermissionsUserMapping{},
			args:              &db.GrantPendingPermissionsArgs{},
			expectCalledCount: 0,
		},
		{
			name: "grant by emails",
			config: &schema.PermissionsUserMapping{
				BindID: "email",
			},
			args: &db.GrantPendingPermissionsArgs{
				UserID: 1,
			},
			mockEmails: []*db.UserEmail{
				{Email: "alice@example.com"},
				{Email: "alice2@example.com"},
			},
			expectCalledCount: 2,
		},
		{
			name: "grant by username",
			config: &schema.PermissionsUserMapping{
				BindID: "username",
			},
			args: &db.GrantPendingPermissionsArgs{
				UserID: 1,
			},
			mockUser: &types.User{
				Username: "alice",
			},
			expectCalledCount: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			globals.SetPermissionsUserMapping(test.config)

			calledCount := 0
			db.Mocks.UserEmails.ListByUser = func(_ context.Context, opt db.UserEmailsListOptions) ([]*db.UserEmail, error) {
				if opt.UserID <= 0 {
					return nil, errors.New("opt.UserID must be greater than 0")
				} else if !opt.OnlyVerified {
					return nil, errors.New("opt.OnlyVerified is not set to true")
				}
				return test.mockEmails, nil
			}
			db.Mocks.Users.GetByID = func(context.Context, int32) (*types.User, error) {
				return test.mockUser, nil
			}
			Mocks.Perms.GrantPendingPermissionsTx = func(context.Context, int32, *iauthz.UserPendingPermissions) error {
				calledCount++
				return nil
			}
			defer func() {
				db.Mocks.UserEmails.ListByUser = nil
				db.Mocks.Users.GetByID = nil
				Mocks.Perms.GrantPendingPermissionsTx = nil
			}()

			err := s.GrantPendingPermissions(ctx, test.args)
			if err != nil {
				t.Fatal(err)
			}

			if test.expectCalledCount != calledCount {
				t.Fatalf("calledCount: want %d but got %d", test.expectCalledCount, calledCount)
			}
		})
	}
}

func TestAuthzStore_AuthorizedRepos(t *testing.T) {
	ctx := context.Background()
	s := NewAuthzStore(nil, clock)

	tests := []struct {
		name        string
		args        *db.AuthorizedReposArgs
		mockPerms   *iauthz.UserPermissions
		expectRepos []*types.Repo
	}{
		{
			name: "no repos",
			args: &db.AuthorizedReposArgs{},
		},
		{
			name: "has repos",
			args: &db.AuthorizedReposArgs{
				Repos: []*types.Repo{
					{ID: 1},
					{ID: 2},
					{ID: 4},
				},
			},
			mockPerms: &iauthz.UserPermissions{
				Type: authz.PermRepos,
				IDs:  toBitmap(1, 2, 3),
			},
			expectRepos: []*types.Repo{
				{ID: 1},
				{ID: 2},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mocks.Perms.LoadUserPermissions = func(_ context.Context, p *iauthz.UserPermissions) error {
				*p = *test.mockPerms
				return nil
			}
			defer func() {
				Mocks.Perms.LoadUserPermissions = nil
			}()

			repos, err := s.AuthorizedRepos(ctx, test.args)
			if err != nil {
				t.Fatal(err)
			}

			equal(t, "repos", test.expectRepos, repos)
		})
	}
}
