package resolvers

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/assert"

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
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var now = timeutil.Now().UnixNano()

func clock() time.Time {
	return time.Unix(0, atomic.LoadInt64(&now))
}

func mustParseGraphQLSchema(t *testing.T, db database.DB) *graphql.Schema {
	t.Helper()

	parsedSchema, err := graphqlbackend.NewSchema(db, nil, nil, nil, NewResolver(db, clock), nil, nil, nil, nil, nil, nil, nil)
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
		gqlTests           func(database.DB) []*graphqlbackend.Test
		expUserIDs         map[int32]struct{}
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
		gqlTests: func(db database.DB) []*graphqlbackend.Test {
			return []*graphqlbackend.Test{{
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
		expUserIDs: map[int32]struct{}{1: {}},
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
		gqlTests: func(db database.DB) []*graphqlbackend.Test {
			return []*graphqlbackend.Test{{
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
		expUserIDs: map[int32]struct{}{1: {}},
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
				ids := p.UserIDs
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

			graphqlbackend.RunTests(t, test.gqlTests(db))
		})
	}
}

func TestResolver_SetRepositoryPermissionsUnrestricted(t *testing.T) {
	// TODO: Factor out this common check
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

	var haveIDs []int32
	var haveUnrestricted bool

	perms := edb.NewMockPermsStore()
	perms.SetRepoPermissionsUnrestrictedFunc.SetDefaultHook(func(ctx context.Context, ids []int32, unrestricted bool) error {
		haveIDs = ids
		haveUnrestricted = unrestricted
		return nil
	})
	users := database.NewStrictMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	db := edb.NewStrictMockEnterpriseDB()
	db.PermsFunc.SetDefaultReturn(perms)
	db.UsersFunc.SetDefaultReturn(users)

	gqlTests := []*graphqlbackend.Test{{
		Schema: mustParseGraphQLSchema(t, db),
		Query: `
						mutation {
							setRepositoryPermissionsUnrestricted(
								repositories: ["UmVwb3NpdG9yeTox","UmVwb3NpdG9yeToy","UmVwb3NpdG9yeToz"],
								unrestricted: true
								) {
								alwaysNil
							}
						}
					`,
		ExpectedResult: `
						{
							"setRepositoryPermissionsUnrestricted": {
								"alwaysNil": null
							}
						}
					`,
	}}

	graphqlbackend.RunTests(t, gqlTests)

	assert.Equal(t, haveIDs, []int32{1, 2, 3})
	assert.True(t, haveUnrestricted)
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
		p.IDs = map[int32]struct{}{1: {}}
		return nil
	})
	perms.LoadUserPendingPermissionsFunc.SetDefaultHook(func(_ context.Context, p *authz.UserPendingPermissions) error {
		p.IDs = map[int32]struct{}{2: {}, 3: {}, 4: {}, 5: {}}
		return nil
	})

	db := edb.NewStrictMockEnterpriseDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(repos)
	db.PermsFunc.SetDefaultReturn(perms)

	tests := []struct {
		name     string
		gqlTests []*graphqlbackend.Test
	}{
		{
			name: "check authorized repos via email",
			gqlTests: []*graphqlbackend.Test{
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
					ExpectedResult: fmt.Sprintf(`
				{
					"authorizedUserRepositories": {
						"nodes": [
							{"id":"%s"}
						]
    				}
				}
			`, graphqlbackend.MarshalRepositoryID(1)),
				},
			},
		},
		{
			name: "check authorized repos via username",
			gqlTests: []*graphqlbackend.Test{
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
					ExpectedResult: fmt.Sprintf(`
				{
					"authorizedUserRepositories": {
						"nodes": [
							{"id":"%s"}
						]
    				}
				}
			`, graphqlbackend.MarshalRepositoryID(1)),
				},
			},
		},
		{
			name: "check pending authorized repos via email",
			gqlTests: []*graphqlbackend.Test{
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
					ExpectedResult: fmt.Sprintf(`
				{
					"authorizedUserRepositories": {
						"nodes": [
							{"id":"%s"},{"id":"%s"},{"id":"%s"},{"id":"%s"}
						]
    				}
				}
			`, graphqlbackend.MarshalRepositoryID(2), graphqlbackend.MarshalRepositoryID(3), graphqlbackend.MarshalRepositoryID(4), graphqlbackend.MarshalRepositoryID(5)),
				},
			},
		},
		{
			name: "check pending authorized repos via username",
			gqlTests: []*graphqlbackend.Test{
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
					ExpectedResult: fmt.Sprintf(`
				{
					"authorizedUserRepositories": {
						"nodes": [
							{"id":"%s"},{"id":"%s"},{"id":"%s"},{"id":"%s"}
						]
    				}
				}
			`, graphqlbackend.MarshalRepositoryID(2), graphqlbackend.MarshalRepositoryID(3), graphqlbackend.MarshalRepositoryID(4), graphqlbackend.MarshalRepositoryID(5)),
				},
			},
		},
		{
			name: "check pending authorized repos via username with pagination, page 1",
			gqlTests: []*graphqlbackend.Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: fmt.Sprintf(`
				{
					authorizedUserRepositories(
						first: 2,
						after: "%s",
						username: "bob") {
						nodes {
							id
						}
					}
				}
			`, graphqlbackend.MarshalRepositoryID(2)),
					ExpectedResult: fmt.Sprintf(`
				{
					"authorizedUserRepositories": {
						"nodes": [
							{"id":"%s"},{"id":"%s"}
						]
                    }
				}
			`, graphqlbackend.MarshalRepositoryID(3), graphqlbackend.MarshalRepositoryID(4)),
				},
			},
		},
		{
			name: "check pending authorized repos via username with pagination, page 2",
			gqlTests: []*graphqlbackend.Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: fmt.Sprintf(`
				{
					authorizedUserRepositories(
						first: 2,
						after: "%s",
						username: "bob") {
						nodes {
							id
						}
					}
				}
			`, graphqlbackend.MarshalRepositoryID(4)),
					ExpectedResult: fmt.Sprintf(`
				{
					"authorizedUserRepositories": {
						"nodes": [
							{"id":"%s"}
						]
    				}
				}
			`, graphqlbackend.MarshalRepositoryID(5)),
				},
			},
		},
		{
			name: "check pending authorized repos via username given no IDs after, after ID, return empty",
			gqlTests: []*graphqlbackend.Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: fmt.Sprintf(`
						{
							authorizedUserRepositories(
								first: 2,
								after: "%s",
								username: "bob") {
								nodes {
									id
								}
							}
						}
					`, graphqlbackend.MarshalRepositoryID(5)),
					ExpectedResult: `
				{
					"authorizedUserRepositories": {
						"nodes": []
    				}
				}
			`,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			graphqlbackend.RunTests(t, test.gqlTests)
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
		gqlTests []*graphqlbackend.Test
	}{
		{
			name: "list pending users with their bind IDs",
			gqlTests: []*graphqlbackend.Test{
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
			graphqlbackend.RunTests(t, test.gqlTests)
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
		p.UserIDs = map[int32]struct{}{1: {}, 2: {}, 3: {}, 4: {}, 5: {}}
		return nil
	})

	db := edb.NewStrictMockEnterpriseDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(repos)
	db.PermsFunc.SetDefaultReturn(perms)

	tests := []struct {
		name     string
		gqlTests []*graphqlbackend.Test
	}{
		{
			name: "get authorized users",
			gqlTests: []*graphqlbackend.Test{
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
					ExpectedResult: fmt.Sprintf(`
				{
					"repository": {
						"authorizedUsers": {
							"nodes":[
								{"id":"%s"},{"id":"%s"},{"id":"%s"},{"id":"%s"},{"id":"%s"}
							]
						}
    				}
				}
			`, graphqlbackend.MarshalUserID(1), graphqlbackend.MarshalUserID(2), graphqlbackend.MarshalUserID(3), graphqlbackend.MarshalUserID(4), graphqlbackend.MarshalUserID(5)),
				},
			},
		},
		{
			name: "get authorized users with pagination, page 1",
			gqlTests: []*graphqlbackend.Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: fmt.Sprintf(`
{
					repository(name: "github.com/owner/repo") {
						authorizedUsers(
							first: 2,
							after: "%s") {
							nodes {
								id
							}
						}
					}
				}
			`, graphqlbackend.MarshalUserID(1)),
					ExpectedResult: fmt.Sprintf(`
				{
					"repository": {
						"authorizedUsers": {
							"nodes":[
								{"id":"%s"},{"id":"%s"}
							]
						}
    				}
				}
			`, graphqlbackend.MarshalUserID(2), graphqlbackend.MarshalUserID(3)),
				},
			},
		},
		{
			name: "get authorized users with pagination, page 2",
			gqlTests: []*graphqlbackend.Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: fmt.Sprintf(`
{
					repository(name: "github.com/owner/repo") {
						authorizedUsers(
							first: 2,
							after: "%s") {
							nodes {
								id
							}
						}
					}
				}
			`, graphqlbackend.MarshalUserID(3)),
					ExpectedResult: fmt.Sprintf(`
				{
					"repository": {
						"authorizedUsers": {
							"nodes":[
								{"id":"%s"},{"id":"%s"}
							]
						}
    				}
				}
			`, graphqlbackend.MarshalUserID(4), graphqlbackend.MarshalUserID(5)),
				},
			},
		},
		{
			name: "get authorized users given no IDs after, after ID, return empty",
			gqlTests: []*graphqlbackend.Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: fmt.Sprintf(`
{
					repository(name: "github.com/owner/repo") {
						authorizedUsers(
							first: 2,
							after: "%s") {
							nodes {
								id
							}
						}
					}
				}
			`, graphqlbackend.MarshalUserID(5)),
					ExpectedResult: `
				{
					"repository": {
						"authorizedUsers": {
							"nodes":[]
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
			graphqlbackend.RunTests(t, test.gqlTests)
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
		gqlTests []*graphqlbackend.Test
	}{
		{
			name: "get permissions information",
			gqlTests: []*graphqlbackend.Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
				{
					repository(name: "github.com/owner/repo") {
						permissionsInfo {
							permissions
							syncedAt
							updatedAt
							unrestricted
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
							"updatedAt": "%[1]s",
							"unrestricted": false
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
			graphqlbackend.RunTests(t, test.gqlTests)
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
		gqlTests []*graphqlbackend.Test
	}{
		{
			name: "get permissions information",
			gqlTests: []*graphqlbackend.Test{
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
			graphqlbackend.RunTests(t, test.gqlTests)
		})
	}
}

func TestResolver_SetSubRepositoryPermissionsForUsers(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewStrictMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)

		subrepos := database.NewStrictMockSubRepoPermsStore()
		subrepos.UpsertFunc.SetDefaultHook(func(ctx context.Context, i int32, id api.RepoID, permissions authz.SubRepoPermissions) error {
			return nil
		})

		db := edb.NewStrictMockEnterpriseDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.SubRepoPermsFunc.SetDefaultReturn(subrepos)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&Resolver{db: db}).SetSubRepositoryPermissionsForUsers(ctx, &graphqlbackend.SubRepoPermsArgs{})
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	t.Run("set sub-repo perms", func(t *testing.T) {
		usersStore := database.NewStrictMockUserStore()
		usersStore.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{
			ID:        1,
			SiteAdmin: true,
		}, nil)
		usersStore.GetByUsernameFunc.SetDefaultHook(func(ctx context.Context, s string) (*types.User, error) {
			return &types.User{
				ID:       1,
				Username: "foo",
			}, nil
		})
		usersStore.GetByVerifiedEmailFunc.SetDefaultHook(func(ctx context.Context, s string) (*types.User, error) {
			return &types.User{
				ID:       1,
				Username: "foo",
			}, nil
		})

		subReposStore := database.NewStrictMockSubRepoPermsStore()
		subReposStore.UpsertFunc.SetDefaultHook(func(ctx context.Context, i int32, id api.RepoID, permissions authz.SubRepoPermissions) error {
			return nil
		})

		reposStore := database.NewStrictMockRepoStore()
		reposStore.GetFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (*types.Repo, error) {
			return &types.Repo{
				ID:   1,
				Name: "foo",
			}, nil
		})

		db := edb.NewStrictMockEnterpriseDB()
		db.TransactFunc.SetDefaultHook(func(ctx context.Context) (database.DB, error) {
			return db, nil
		})
		db.DoneFunc.SetDefaultHook(func(err error) error {
			return nil
		})
		db.UsersFunc.SetDefaultReturn(usersStore)
		db.SubRepoPermsFunc.SetDefaultReturn(subReposStore)
		db.ReposFunc.SetDefaultReturn(reposStore)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

		test := &graphqlbackend.Test{
			Context: ctx,
			Schema:  mustParseGraphQLSchema(t, db),
			Query: `
						mutation {
  setSubRepositoryPermissionsForUsers(
    repository: "UmVwb3NpdG9yeTox"
    userPermissions: [{bindID: "alice", pathIncludes: ["*"], pathExcludes: ["*_test.go"]}]
  ) {
    alwaysNil
  }
}
					`,
			ExpectedResult: `
						{
							"setSubRepositoryPermissionsForUsers": {
								"alwaysNil": null
							}
						}
					`,
		}

		graphqlbackend.RunTests(t, []*graphqlbackend.Test{test})

		// Assert that we actually tried to store perms
		h := subReposStore.UpsertFunc.History()
		if len(h) != 1 {
			t.Fatalf("Wanted 1 call, got %d", len(h))
		}
	})
}
