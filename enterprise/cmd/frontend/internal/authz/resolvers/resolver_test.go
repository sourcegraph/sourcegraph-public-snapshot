package resolvers

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	edb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/db"
	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

var now = time.Now().Truncate(time.Microsecond).UnixNano()

func clock() time.Time {
	return time.Unix(0, atomic.LoadInt64(&now)).Truncate(time.Microsecond)
}

func mustParseGraphQLSchema(t *testing.T, db *sql.DB) *graphql.Schema {
	t.Helper()

	schema, err := graphqlbackend.NewSchema(nil, nil, NewResolver(db, clock))
	if err != nil {
		t.Fatal(err)
	}

	return schema
}

func TestResolver_SetRepositoryPermissionsForUsers(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{}, nil
		}
		defer func() {
			db.Mocks.Users.GetByCurrentAuthUser = nil
		}()

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&Resolver{}).SetRepositoryPermissionsForUsers(ctx, &graphqlbackend.RepoPermsArgs{})
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	db.Mocks.Repos.Get = func(_ context.Context, id api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: id}, nil
	}
	defer func() {
		db.Mocks.Users.GetByCurrentAuthUser = nil
		db.Mocks.Repos.Get = nil
	}()

	tests := []struct {
		name                 string
		config               *schema.PermissionsUserMapping
		mockVerifiedEmails   []*db.UserEmail
		mockUsers            []*types.User
		gqlTests             []*gqltesting.Test
		expectUserIDs        []uint32
		expectPendingBindIDs []string
	}{
		{
			name: "set permissions via email",
			config: &schema.PermissionsUserMapping{
				BindID: "email",
			},
			mockVerifiedEmails: []*db.UserEmail{
				{
					UserID: 1,
					Email:  "alice@example.com",
				},
			},
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, nil),
					Query: `
				mutation {
					setRepositoryPermissionsForUsers(
						repository: "UmVwb3NpdG9yeTox",
						bindIDs: ["alice@example.com", "bob"]) {
						alwaysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"setRepositoryPermissionsForUsers": {
						"alwaysNil": null
    				}
				}
			`,
				},
			},
			expectUserIDs:        []uint32{1},
			expectPendingBindIDs: []string{"bob"},
		},
		{
			name: "set permissions via username",
			config: &schema.PermissionsUserMapping{
				BindID: "username",
			},
			mockUsers: []*types.User{
				{
					ID:       1,
					Username: "alice",
				},
			},
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, nil),
					Query: `
				mutation {
					setRepositoryPermissionsForUsers(
						repository: "UmVwb3NpdG9yeTox",
						bindIDs: ["alice", "bob"]) {
						alwaysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"setRepositoryPermissionsForUsers": {
						"alwaysNil": null
    				}
				}
			`,
				},
			},
			expectUserIDs:        []uint32{1},
			expectPendingBindIDs: []string{"bob"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			globals.SetPermissionsUserMapping(test.config)

			db.Mocks.UserEmails.GetVerifiedEmails = func(context.Context, ...string) ([]*db.UserEmail, error) {
				return test.mockVerifiedEmails, nil
			}
			db.Mocks.Users.GetByUsernames = func(context.Context, ...string) ([]*types.User, error) {
				return test.mockUsers, nil
			}
			edb.Mocks.Perms.SetRepoPermissions = func(_ context.Context, p *iauthz.RepoPermissions) error {
				ids := p.UserIDs.ToArray()
				if diff := cmp.Diff(test.expectUserIDs, ids); diff != "" {
					return fmt.Errorf("p.UserIDs: %v", diff)
				}
				return nil
			}
			edb.Mocks.Perms.SetRepoPendingPermissions = func(_ context.Context, bindIDs []string, _ *iauthz.RepoPermissions) error {
				if diff := cmp.Diff(test.expectPendingBindIDs, bindIDs); diff != "" {
					return fmt.Errorf("bindIDs: %v", diff)
				}
				return nil
			}
			defer func() {
				db.Mocks.UserEmails.GetVerifiedEmails = nil
				db.Mocks.Users.GetByUsernames = nil
				edb.Mocks.Perms.SetRepoPermissions = nil
				edb.Mocks.Perms.SetRepoPendingPermissions = nil
			}()

			gqltesting.RunTests(t, test.gqlTests)
		})
	}
}

func TestResolver_AuthorizedUserRepositories(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{}, nil
		}
		defer func() {
			db.Mocks.Users.GetByCurrentAuthUser = nil
		}()

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&Resolver{}).AuthorizedUserRepositories(ctx, &graphqlbackend.AuthorizedRepoArgs{})
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	db.Mocks.Users.GetByVerifiedEmail = func(_ context.Context, email string) (*types.User, error) {
		if email == "alice@example.com" {
			return &types.User{ID: 1}, nil
		}
		return nil, db.MockUserNotFoundErr
	}
	db.Mocks.Users.GetByUsername = func(_ context.Context, username string) (*types.User, error) {
		if username == "alice" {
			return &types.User{ID: 1}, nil
		}
		return nil, db.MockUserNotFoundErr
	}
	db.Mocks.Repos.GetByIDs = func(_ context.Context, ids ...api.RepoID) ([]*types.Repo, error) {
		repos := make([]*types.Repo, len(ids))
		for i, id := range ids {
			repos[i] = &types.Repo{ID: id}
		}
		return repos, nil
	}
	edb.Mocks.Perms.LoadUserPermissions = func(_ context.Context, p *iauthz.UserPermissions) error {
		p.IDs = roaring.NewBitmap()
		p.IDs.Add(1)
		return nil
	}
	edb.Mocks.Perms.LoadUserPendingPermissions = func(_ context.Context, p *iauthz.UserPendingPermissions) error {
		p.IDs = roaring.NewBitmap()
		p.IDs.Add(2)
		return nil
	}
	defer func() {
		db.Mocks.Users.GetByCurrentAuthUser = nil
		db.Mocks.Users.GetByVerifiedEmail = nil
		db.Mocks.Users.GetByUsername = nil
		edb.Mocks.Perms.LoadUserPermissions = nil
		edb.Mocks.Perms.LoadUserPendingPermissions = nil
	}()

	tests := []struct {
		name     string
		gqlTests []*gqltesting.Test
	}{
		{
			name: "check authorized repos via email",
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, nil),
					Query: `
				{
					authorizedUserRepositories(
						first: 10,
						email: "alice@example.com") {
						nodes {
							id
						}
					}
				}
			`,
					ExpectedResult: `
				{
					"authorizedUserRepositories": {
						"nodes": [
							{"id":"UmVwb3NpdG9yeTox"}
						]
    				}
				}
			`,
				},
			},
		},
		{
			name: "check authorized repos via username",
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, nil),
					Query: `
				{
					authorizedUserRepositories(
						first: 10,
						username: "alice") {
						nodes {
							id
						}
					}
				}
			`,
					ExpectedResult: `
				{
					"authorizedUserRepositories": {
						"nodes": [
							{"id":"UmVwb3NpdG9yeTox"}
						]
    				}
				}
			`,
				},
			},
		},
		{
			name: "check pending authorized repos via email",
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, nil),
					Query: `
				{
					authorizedUserRepositories(
						first: 10,
						email: "bob@example.com") {
						nodes {
							id
						}
					}
				}
			`,
					ExpectedResult: `
				{
					"authorizedUserRepositories": {
						"nodes": [
							{"id":"UmVwb3NpdG9yeToy"}
						]
    				}
				}
			`,
				},
			},
		},
		{
			name: "check pending authorized repos via username",
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, nil),
					Query: `
				{
					authorizedUserRepositories(
						first: 10,
						username: "bob") {
						nodes {
							id
						}
					}
				}
			`,
					ExpectedResult: `
				{
					"authorizedUserRepositories": {
						"nodes": [
							{"id":"UmVwb3NpdG9yeToy"}
						]
    				}
				}
			`,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gqltesting.RunTests(t, test.gqlTests)
		})
	}
}

func TestResolver_UsersWithPendingPermissions(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{}, nil
		}
		defer func() {
			db.Mocks.Users.GetByCurrentAuthUser = nil
		}()

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&Resolver{}).UsersWithPendingPermissions(ctx)
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	edb.Mocks.Perms.ListPendingUsers = func(context.Context) ([]string, error) {
		return []string{"alice", "bob"}, nil
	}
	defer func() {
		db.Mocks.Users.GetByCurrentAuthUser = nil
		edb.Mocks.Perms.ListPendingUsers = nil
	}()

	tests := []struct {
		name     string
		gqlTests []*gqltesting.Test
	}{
		{
			name: "list pending users with their bind IDs",
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, nil),
					Query: `
				{
					usersWithPendingPermissions
				}
			`,
					ExpectedResult: `
				{
					"usersWithPendingPermissions": [
						"alice",
						"bob"
					]
				}
			`,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gqltesting.RunTests(t, test.gqlTests)
		})
	}
}

func TestResolver_AuthorizedUsers(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{}, nil
		}
		defer func() {
			db.Mocks.Users.GetByCurrentAuthUser = nil
		}()

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&Resolver{}).AuthorizedUsers(ctx, &graphqlbackend.RepoAuthorizedUserArgs{})
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	db.Mocks.Users.List = func(_ context.Context, opt *db.UsersListOptions) ([]*types.User, error) {
		users := make([]*types.User, len(opt.UserIDs))
		for i, id := range opt.UserIDs {
			users[i] = &types.User{ID: id}
		}
		return users, nil
	}
	db.Mocks.Repos.GetByName = func(_ context.Context, repo api.RepoName) (*types.Repo, error) {
		return &types.Repo{ID: 1, Name: repo}, nil
	}
	db.Mocks.Repos.Get = func(_ context.Context, id api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: id}, nil
	}
	edb.Mocks.Perms.LoadRepoPermissions = func(_ context.Context, p *iauthz.RepoPermissions) error {
		p.UserIDs = roaring.NewBitmap()
		p.UserIDs.Add(1)
		return nil
	}
	defer func() {
		db.Mocks.Users.GetByCurrentAuthUser = nil
		db.Mocks.Users.List = nil
		db.Mocks.Repos.GetByName = nil
		db.Mocks.Repos.Get = nil
		edb.Mocks.Perms.LoadRepoPermissions = nil
	}()

	tests := []struct {
		name     string
		gqlTests []*gqltesting.Test
	}{
		{
			name: "check authorized repos via email",
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, nil),
					Query: `
				{
					repository(name: "github.com/owner/repo") {
						authorizedUsers(first: 10) {
							nodes {
								id
							}
						}
					}
				}
			`,
					ExpectedResult: `
				{
					"repository": {
						"authorizedUsers": {
							"nodes":[
								{"id":"VXNlcjox"}
							]
						}
    				}
				}
			`,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gqltesting.RunTests(t, test.gqlTests)
		})
	}
}
