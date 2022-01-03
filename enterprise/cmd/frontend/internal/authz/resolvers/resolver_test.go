package resolvers

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/gqltesting"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

var now = timeutil.Now().UnixNano()

func clock() time.Time {
	return time.Unix(0, atomic.LoadInt64(&now))
}

func mustParseGraphQLSchema(t *testing.T, db database.DB) *graphql.Schema {
	t.Helper()

	parsedSchema, err := graphqlbackend.NewSchema(db, nil, nil, nil, NewResolver(db, clock), nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	return parsedSchema
}

func TestResolver_SetRepositoryPermissionsForUsers(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewStrictMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)

		db := edb.NewStrictMockEnterpriseDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&Resolver{db: db}).SetRepositoryPermissionsForUsers(ctx, &graphqlbackend.RepoPermsArgs{})
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	tests := []struct {
		name               string
		config             *schema.PermissionsUserMapping
		mockVerifiedEmails []*database.UserEmail
		mockUsers          []*types.User
		gqlTests           func(database.DB) []*gqltesting.Test
		expUserIDs         []uint32
		expAccounts        *extsvc.Accounts
	}{{
		name: "set permissions via email",
		config: &schema.PermissionsUserMapping{
			BindID: "email",
		},
		mockVerifiedEmails: []*database.UserEmail{
			{
				UserID: 1,
				Email:  "alice@example.com",
			},
		},
		gqlTests: func(db database.DB) []*gqltesting.Test {
			return []*gqltesting.Test{{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
							mutation {
								setRepositoryPermissionsForUsers(
									repository: "UmVwb3NpdG9yeTox",
									userPermissions: [
										{ bindID: "alice@example.com"},
										{ bindID: "bob"}
									]) {
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
			}
		},
		expUserIDs: []uint32{1},
		expAccounts: &extsvc.Accounts{
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			AccountIDs:  []string{"bob"},
		},
	}, {
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
		gqlTests: func(db database.DB) []*gqltesting.Test {
			return []*gqltesting.Test{{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
						mutation {
							setRepositoryPermissionsForUsers(
								repository: "UmVwb3NpdG9yeTox",
								userPermissions: [
									{ bindID: "alice"},
									{ bindID: "bob"}
								]) {
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
			}}
		},
		expUserIDs: []uint32{1},
		expAccounts: &extsvc.Accounts{
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			AccountIDs:  []string{"bob"},
		},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			globals.SetPermissionsUserMapping(test.config)

			users := database.NewStrictMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
			users.GetByUsernamesFunc.SetDefaultReturn(test.mockUsers, nil)

			userEmails := database.NewStrictMockUserEmailsStore()
			userEmails.GetVerifiedEmailsFunc.SetDefaultReturn(test.mockVerifiedEmails, nil)

			repos := database.NewStrictMockRepoStore()
			repos.GetFunc.SetDefaultHook(func(_ context.Context, id api.RepoID) (*types.Repo, error) {
				return &types.Repo{ID: id}, nil
			})

			perms := edb.NewStrictMockPermsStore()
			perms.TransactFunc.SetDefaultReturn(perms, nil)
			perms.DoneFunc.SetDefaultReturn(nil)
			perms.SetRepoPermissionsFunc.SetDefaultHook(func(_ context.Context, p *authz.RepoPermissions) error {
				ids := p.UserIDs.ToArray()
				if diff := cmp.Diff(test.expUserIDs, ids); diff != "" {
					return errors.Errorf("p.UserIDs: %v", diff)
				}
				return nil
			})
			perms.SetRepoPendingPermissionsFunc.SetDefaultHook(func(_ context.Context, accounts *extsvc.Accounts, _ *authz.RepoPermissions) error {
				if diff := cmp.Diff(test.expAccounts, accounts); diff != "" {
					return errors.Errorf("accounts: %v", diff)
				}
				return nil
			})

			db := edb.NewStrictMockEnterpriseDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.UserEmailsFunc.SetDefaultReturn(userEmails)
			db.ReposFunc.SetDefaultReturn(repos)
			db.PermsFunc.SetDefaultReturn(perms)

			gqltesting.RunTests(t, test.gqlTests(db))
		})
	}
}

func TestResolver_ScheduleRepositoryPermissionsSync(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewStrictMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)

		db := edb.NewStrictMockEnterpriseDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&Resolver{db: db}).ScheduleRepositoryPermissionsSync(ctx, &graphqlbackend.RepositoryIDArgs{})
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	users := database.NewStrictMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	db := edb.NewStrictMockEnterpriseDB()
	db.UsersFunc.SetDefaultReturn(users)

	r := &Resolver{
		db: db,
		repoupdaterClient: &fakeRepoupdaterClient{
			mockSchedulePermsSync: func(ctx context.Context, args protocol.PermsSyncRequest) error {
				if len(args.RepoIDs) != 1 {
					return errors.Errorf("RepoIDs: want 1 id but got %d", len(args.RepoIDs))
				}
				return nil
			},
		},
	}
	_, err := r.ScheduleRepositoryPermissionsSync(context.Background(), &graphqlbackend.RepositoryIDArgs{
		Repository: graphqlbackend.MarshalRepositoryID(1),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestResolver_ScheduleUserPermissionsSync(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewStrictMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)

		db := edb.NewStrictMockEnterpriseDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&Resolver{db: db}).ScheduleUserPermissionsSync(ctx, &graphqlbackend.UserPermissionsSyncArgs{})
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	users := database.NewStrictMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	db := edb.NewStrictMockEnterpriseDB()
	db.UsersFunc.SetDefaultReturn(users)

	t.Run("queue a user", func(t *testing.T) {
		r := &Resolver{
			db: db,
			repoupdaterClient: &fakeRepoupdaterClient{
				mockSchedulePermsSync: func(ctx context.Context, args protocol.PermsSyncRequest) error {
					if len(args.UserIDs) != 1 {
						return errors.Errorf("UserIDs: want 1 id but got %d", len(args.UserIDs))
					}
					return nil
				},
			},
		}
		_, err := r.ScheduleUserPermissionsSync(context.Background(), &graphqlbackend.UserPermissionsSyncArgs{
			User: graphqlbackend.MarshalUserID(1),
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("queue a user with options", func(t *testing.T) {
		r := &Resolver{
			db: db,
			repoupdaterClient: &fakeRepoupdaterClient{
				mockSchedulePermsSync: func(ctx context.Context, args protocol.PermsSyncRequest) error {
					if len(args.UserIDs) != 1 {
						return errors.Errorf("UserIDs: want 1 id but got %d", len(args.UserIDs))
					}
					if !args.Options.InvalidateCaches {
						return errors.Errorf("Options.InvalidateCaches: expected true but got false")
					}
					return nil
				},
			},
		}
		trueVal := true
		_, err := r.ScheduleUserPermissionsSync(context.Background(), &graphqlbackend.UserPermissionsSyncArgs{
			User:    graphqlbackend.MarshalUserID(1),
			Options: &struct{ InvalidateCaches *bool }{InvalidateCaches: &trueVal},
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

type fakeRepoupdaterClient struct {
	mockSchedulePermsSync func(ctx context.Context, args protocol.PermsSyncRequest) error
}

func (c *fakeRepoupdaterClient) SchedulePermsSync(ctx context.Context, args protocol.PermsSyncRequest) error {
	return c.mockSchedulePermsSync(ctx, args)
}

func TestResolver_AuthorizedUserRepositories(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewStrictMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)

		db := edb.NewStrictMockEnterpriseDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&Resolver{db: db}).AuthorizedUserRepositories(ctx, &graphqlbackend.AuthorizedRepoArgs{})
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	users := database.NewStrictMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	users.GetByVerifiedEmailFunc.SetDefaultHook(func(_ context.Context, email string) (*types.User, error) {
		if email == "alice@example.com" {
			return &types.User{ID: 1}, nil
		}
		return nil, database.MockUserNotFoundErr
	})
	users.GetByUsernameFunc.SetDefaultHook(func(_ context.Context, username string) (*types.User, error) {
		if username == "alice" {
			return &types.User{ID: 1}, nil
		}
		return nil, database.MockUserNotFoundErr
	})

	repos := database.NewStrictMockRepoStore()
	repos.GetByIDsFunc.SetDefaultHook(func(_ context.Context, ids ...api.RepoID) ([]*types.Repo, error) {
		repos := make([]*types.Repo, len(ids))
		for i, id := range ids {
			repos[i] = &types.Repo{ID: id}
		}
		return repos, nil
	})

	perms := edb.NewStrictMockPermsStore()
	perms.LoadUserPermissionsFunc.SetDefaultHook(func(_ context.Context, p *authz.UserPermissions) error {
		p.IDs = roaring.NewBitmap()
		p.IDs.Add(1)
		return nil
	})
	perms.LoadUserPendingPermissionsFunc.SetDefaultHook(func(_ context.Context, p *authz.UserPendingPermissions) error {
		p.IDs = roaring.NewBitmap()
		p.IDs.Add(2)
		return nil
	})

	db := edb.NewStrictMockEnterpriseDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(repos)
	db.PermsFunc.SetDefaultReturn(perms)

	tests := []struct {
		name     string
		gqlTests []*gqltesting.Test
	}{
		{
			name: "check authorized repos via email",
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
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
					Schema: mustParseGraphQLSchema(t, db),
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
					Schema: mustParseGraphQLSchema(t, db),
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
					Schema: mustParseGraphQLSchema(t, db),
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
		users := database.NewStrictMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)

		db := edb.NewStrictMockEnterpriseDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&Resolver{db: db}).UsersWithPendingPermissions(ctx)
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	users := database.NewStrictMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	perms := edb.NewStrictMockPermsStore()
	perms.ListPendingUsersFunc.SetDefaultReturn([]string{"alice", "bob"}, nil)

	db := edb.NewStrictMockEnterpriseDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.PermsFunc.SetDefaultReturn(perms)

	tests := []struct {
		name     string
		gqlTests []*gqltesting.Test
	}{
		{
			name: "list pending users with their bind IDs",
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
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
		users := database.NewStrictMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)

		db := edb.NewStrictMockEnterpriseDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&Resolver{db: db}).AuthorizedUsers(ctx, &graphqlbackend.RepoAuthorizedUserArgs{})
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	users := database.NewStrictMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	users.ListFunc.SetDefaultHook(func(_ context.Context, opt *database.UsersListOptions) ([]*types.User, error) {
		users := make([]*types.User, len(opt.UserIDs))
		for i, id := range opt.UserIDs {
			users[i] = &types.User{ID: id}
		}
		return users, nil
	})

	repos := database.NewStrictMockRepoStore()
	repos.GetByNameFunc.SetDefaultHook(func(_ context.Context, repo api.RepoName) (*types.Repo, error) {
		return &types.Repo{ID: 1, Name: repo}, nil
	})
	repos.GetFunc.SetDefaultHook(func(_ context.Context, id api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: id}, nil
	})

	perms := edb.NewStrictMockPermsStore()
	perms.LoadRepoPermissionsFunc.SetDefaultHook(func(_ context.Context, p *authz.RepoPermissions) error {
		p.UserIDs = roaring.NewBitmap()
		p.UserIDs.Add(1)
		return nil
	})

	db := edb.NewStrictMockEnterpriseDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(repos)
	db.PermsFunc.SetDefaultReturn(perms)

	tests := []struct {
		name     string
		gqlTests []*gqltesting.Test
	}{
		{
			name: "get authorized users",
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
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

func TestResolver_RepositoryPermissionsInfo(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewStrictMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)

		db := edb.NewStrictMockEnterpriseDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&Resolver{db: db}).RepositoryPermissionsInfo(ctx, graphqlbackend.MarshalRepositoryID(1))
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	users := database.NewStrictMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	repos := database.NewStrictMockRepoStore()
	repos.GetByNameFunc.SetDefaultHook(func(_ context.Context, repo api.RepoName) (*types.Repo, error) {
		return &types.Repo{ID: 1, Name: repo}, nil
	})
	repos.GetFunc.SetDefaultHook(func(_ context.Context, id api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: id}, nil
	})

	perms := edb.NewStrictMockPermsStore()
	perms.LoadRepoPermissionsFunc.SetDefaultHook(func(_ context.Context, p *authz.RepoPermissions) error {
		p.UpdatedAt = clock()
		p.SyncedAt = clock()
		return nil
	})

	db := edb.NewStrictMockEnterpriseDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(repos)
	db.PermsFunc.SetDefaultReturn(perms)

	tests := []struct {
		name     string
		gqlTests []*gqltesting.Test
	}{
		{
			name: "get permissions information",
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
				{
					repository(name: "github.com/owner/repo") {
						permissionsInfo {
							permissions
							syncedAt
							updatedAt
						}
					}
				}
			`,
					ExpectedResult: fmt.Sprintf(`
				{
					"repository": {
						"permissionsInfo": {
							"permissions": ["READ"],
							"syncedAt": "%[1]s",
							"updatedAt": "%[1]s"
						}
    				}
				}
			`, clock().Format(time.RFC3339)),
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

func TestResolver_UserPermissionsInfo(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewStrictMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)

		db := edb.NewStrictMockEnterpriseDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&Resolver{db: db}).UserPermissionsInfo(ctx, graphqlbackend.MarshalRepositoryID(1))
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	users := database.NewStrictMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	users.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id}, nil
	})

	perms := edb.NewStrictMockPermsStore()
	perms.LoadUserPermissionsFunc.SetDefaultHook(func(_ context.Context, p *authz.UserPermissions) error {
		p.UpdatedAt = clock()
		p.SyncedAt = clock()
		return nil
	})

	repos := database.NewStrictMockRepoStore()
	repos.GetByNameFunc.SetDefaultHook(func(_ context.Context, name api.RepoName) (*types.Repo, error) {
		return &types.Repo{Name: name}, nil
	})

	db := edb.NewStrictMockEnterpriseDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.PermsFunc.SetDefaultReturn(perms)
	db.ReposFunc.SetDefaultReturn(repos)

	tests := []struct {
		name     string
		gqlTests []*gqltesting.Test
	}{
		{
			name: "get permissions information",
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
				{
					currentUser {
						permissionsInfo {
							permissions
							syncedAt
							updatedAt
						}
					}
				}
			`,
					ExpectedResult: fmt.Sprintf(`
				{
					"currentUser": {
						"permissionsInfo": {
							"permissions": ["READ"],
							"syncedAt": "%[1]s",
							"updatedAt": "%[1]s"
						}
    				}
				}
			`, clock().Format(time.RFC3339)),
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
