package githubapp

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	gqlerrors "github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/github_apps/store"
	ghtypes "github.com/sourcegraph/sourcegraph/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// userCtx returns a context where give user ID identifies logged in user.
func userCtx(userID int32) context.Context {
	ctx := context.Background()
	a := actor.FromUser(userID)
	return actor.WithActor(ctx, a)
}

func TestResolver_DeleteGitHubApp(t *testing.T) {
	logger := logtest.Scoped(t)
	id := 1
	graphqlID := MarshalGitHubAppID(int64(id))

	userStore := dbmocks.NewMockUserStore()
	userStore.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
		a := actor.FromContext(ctx)
		if a.UID == 1 {
			return &types.User{ID: 1, SiteAdmin: true}, nil
		}
		if a.UID == 2 {
			return &types.User{ID: 2, SiteAdmin: false}, nil
		}
		return nil, errors.New("not found")
	})

	gitHubAppsStore := store.NewStrictMockGitHubAppsStore()
	gitHubAppsStore.DeleteFunc.SetDefaultHook(func(ctx context.Context, app int) error {
		if app != id {
			return fmt.Errorf("app with id %d does not exist", id)
		}
		return nil
	})

	db := dbmocks.NewStrictMockDB()

	db.GitHubAppsFunc.SetDefaultReturn(gitHubAppsStore)
	db.UsersFunc.SetDefaultReturn(userStore)

	adminCtx := userCtx(1)
	userCtx := userCtx(2)

	schema, err := graphqlbackend.NewSchema(db, gitserver.NewTestClient(t), []graphqlbackend.OptionalResolver{{GitHubAppsResolver: NewResolver(logger, db)}})
	require.NoError(t, err)

	graphqlbackend.RunTests(t, []*graphqlbackend.Test{{
		Schema:  schema,
		Context: adminCtx,
		Query: `
			mutation DeleteGitHubApp($gitHubApp: ID!) {
				deleteGitHubApp(gitHubApp: $gitHubApp) {
					alwaysNil
				}
			}`,
		ExpectedResult: `{
			"deleteGitHubApp": {
				"alwaysNil": null
			}
		}`,
		Variables: map[string]any{
			"gitHubApp": string(graphqlID),
		},
	}, {
		Schema:  schema,
		Context: userCtx,
		Query: `
			mutation DeleteGitHubApp($gitHubApp: ID!) {
				deleteGitHubApp(gitHubApp: $gitHubApp) {
					alwaysNil
				}
			}`,
		ExpectedResult: `{"deleteGitHubApp": null}`,
		ExpectedErrors: []*gqlerrors.QueryError{{
			Message: "must be site admin",
			Path:    []any{string("deleteGitHubApp")},
		}},
		Variables: map[string]any{
			"gitHubApp": string(graphqlID),
		},
	}})
}

func TestResolver_GitHubApps(t *testing.T) {
	logger := logtest.Scoped(t)
	id := 1
	graphqlID := MarshalGitHubAppID(int64(id))

	id2 := 2
	graphqlID2 := MarshalGitHubAppID(int64(id2))

	userStore := dbmocks.NewMockUserStore()
	userStore.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
		a := actor.FromContext(ctx)
		if a.UID == 1 {
			return &types.User{ID: 1, SiteAdmin: true}, nil
		}
		if a.UID == 2 {
			return &types.User{ID: 2, SiteAdmin: false}, nil
		}
		return nil, errors.New("not found")
	})

	gitHubAppsStore := store.NewStrictMockGitHubAppsStore()
	gitHubAppsStore.ListFunc.SetDefaultHook(func(ctx context.Context, domain *types.GitHubAppDomain) ([]*ghtypes.GitHubApp, error) {
		if domain != nil {
			return []*ghtypes.GitHubApp{{ID: 2}}, nil
		}
		return []*ghtypes.GitHubApp{{ID: 1}, {ID: 2}}, nil
	})

	db := dbmocks.NewStrictMockDB()

	db.GitHubAppsFunc.SetDefaultReturn(gitHubAppsStore)
	db.UsersFunc.SetDefaultReturn(userStore)

	adminCtx := userCtx(1)
	userCtx := userCtx(2)

	schema, err := graphqlbackend.NewSchema(db, gitserver.NewTestClient(t), []graphqlbackend.OptionalResolver{{GitHubAppsResolver: NewResolver(logger, db)}})
	require.NoError(t, err)

	graphqlbackend.RunTests(t, []*graphqlbackend.Test{
		{
			Schema:  schema,
			Context: adminCtx,
			Query: `
				query GitHubApps() {
					gitHubApps {
						nodes {
							id
						}
					}
				}`,
			ExpectedResult: fmt.Sprintf(`{
				"gitHubApps": {
					"nodes": [
						{"id":"%s"},
						{"id":"%s"}
					]
				}
			}`, graphqlID, graphqlID2),
		},
		{
			Schema:  schema,
			Context: adminCtx,
			Query: `
				query GitHubApps($domain: GitHubAppDomain) {
					gitHubApps(domain: $domain) {
						nodes {
							id
						}
					}
				}`,
			Variables: map[string]any{
				"domain": types.ReposGitHubAppDomain.ToGraphQL(),
			},
			ExpectedResult: fmt.Sprintf(`{
				"gitHubApps": {
					"nodes": [
						{"id":"%s"}
					]
				}
			}`, graphqlID2),
		},
		{
			Schema:  schema,
			Context: userCtx,
			Query: `
				query GitHubApps() {
					gitHubApps {
						nodes {
							id
						}
					}
				}`,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{{
				Message: "must be site admin",
				Path:    []any{string("gitHubApps")},
			}},
		},
	})
}

func TestResolver_GitHubApp(t *testing.T) {
	logger := logtest.Scoped(t)
	id := 1
	graphqlID := MarshalGitHubAppID(int64(id))

	userStore := dbmocks.NewMockUserStore()
	userStore.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
		a := actor.FromContext(ctx)
		if a.UID == 1 {
			return &types.User{ID: 1, SiteAdmin: true}, nil
		}
		if a.UID == 2 {
			return &types.User{ID: 2, SiteAdmin: false}, nil
		}
		return nil, errors.New("not found")
	})

	gitHubAppsStore := store.NewStrictMockGitHubAppsStore()
	gitHubAppsStore.GetByIDFunc.SetDefaultReturn(&ghtypes.GitHubApp{
		ID:      1,
		AppID:   1,
		BaseURL: "https://github.com",
	}, nil)

	db := dbmocks.NewStrictMockDB()

	db.GitHubAppsFunc.SetDefaultReturn(gitHubAppsStore)
	db.UsersFunc.SetDefaultReturn(userStore)

	adminCtx := userCtx(1)
	userCtx := userCtx(2)

	schema, err := graphqlbackend.NewSchema(db, gitserver.NewTestClient(t), []graphqlbackend.OptionalResolver{{GitHubAppsResolver: NewResolver(logger, db)}})
	require.NoError(t, err)

	graphqlbackend.RunTests(t, []*graphqlbackend.Test{{
		Schema:  schema,
		Context: adminCtx,
		Query: fmt.Sprintf(`
			query {
				gitHubApp(id: "%s") {
					id
				}
			}`, graphqlID),
		ExpectedResult: fmt.Sprintf(`{
			"gitHubApp": {
				"id": "%s"
			}
		}`, graphqlID),
	}, {
		Schema:  schema,
		Context: userCtx,
		Query: fmt.Sprintf(`
			query {
				gitHubApp(id: "%s") {
					id
				}
			}`, graphqlID),
		ExpectedResult: `{"gitHubApp": null}`,
		ExpectedErrors: []*gqlerrors.QueryError{{
			Message: "must be site admin",
			Path:    []any{string("gitHubApp")},
		}},
	}})
}

func TestResolver_GitHubAppByAppID(t *testing.T) {
	logger := logtest.Scoped(t)
	id := 1
	appID := 1234
	baseURL := "https://github.com"
	name := "Horsegraph App"

	userStore := dbmocks.NewMockUserStore()
	userStore.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
		a := actor.FromContext(ctx)
		if a.UID == 1 {
			return &types.User{ID: 1, SiteAdmin: true}, nil
		}
		if a.UID == 2 {
			return &types.User{ID: 2, SiteAdmin: false}, nil
		}
		return nil, errors.New("not found")
	})

	gitHubAppsStore := store.NewStrictMockGitHubAppsStore()
	gitHubAppsStore.GetByAppIDFunc.SetDefaultReturn(&ghtypes.GitHubApp{
		ID:      id,
		AppID:   appID,
		BaseURL: baseURL,
		Name:    name,
	}, nil)

	db := dbmocks.NewStrictMockDB()

	db.GitHubAppsFunc.SetDefaultReturn(gitHubAppsStore)
	db.UsersFunc.SetDefaultReturn(userStore)

	adminCtx := userCtx(1)
	userCtx := userCtx(2)

	schema, err := graphqlbackend.NewSchema(db, gitserver.NewTestClient(t), []graphqlbackend.OptionalResolver{{GitHubAppsResolver: NewResolver(logger, db)}})
	require.NoError(t, err)

	graphqlbackend.RunTests(t, []*graphqlbackend.Test{{
		Schema:  schema,
		Context: adminCtx,
		Query: fmt.Sprintf(`
			query {
				gitHubAppByAppID(appID: %d, baseURL: "%s") {
					id
					name
				}
			}`, appID, baseURL),
		ExpectedResult: fmt.Sprintf(`{
			"gitHubAppByAppID": {
				"id": "%s",
				"name": "%s"
			}
		}`, MarshalGitHubAppID(int64(id)), name),
	}, {
		Schema:  schema,
		Context: userCtx,
		Query: fmt.Sprintf(`
			query {
				gitHubAppByAppID(appID: %d, baseURL: "%s") {
					id
				}
			}`, appID, baseURL),
		ExpectedResult: `{"gitHubAppByAppID": null}`,
		ExpectedErrors: []*gqlerrors.QueryError{{
			Message: "must be site admin",
			Path:    []any{string("gitHubAppByAppID")},
		}},
	}})
}
