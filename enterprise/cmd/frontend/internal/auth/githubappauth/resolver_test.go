package githubapp

import (
	"context"
	"errors"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	gqlerrors "github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/store"
	ghtypes "github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// userCtx returns a context where give user ID identifies logged in user.
func userCtx(userID int32) context.Context {
	ctx := context.Background()
	a := actor.FromUser(userID)
	return actor.WithActor(ctx, a)
}

func TestResolver_CreateGitHubApp(t *testing.T) {
	logger := logtest.Scoped(t)
	id := 1

	userStore := database.NewMockUserStore()
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
	gitHubAppsStore.CreateFunc.SetDefaultHook(func(ctx context.Context, app *ghtypes.GitHubApp) (int, error) {
		if app == nil {
			return 0, errors.New("app is nil")
		}
		o := *app
		o.ID = id
		id++
		return o.ID, nil
	})

	db := edb.NewStrictMockEnterpriseDB()

	db.GitHubAppsFunc.SetDefaultReturn(gitHubAppsStore)
	db.UsersFunc.SetDefaultReturn(userStore)

	adminCtx := userCtx(1)
	userCtx := userCtx(2)

	schema, err := graphqlbackend.NewSchema(db, gitserver.NewClient(), nil, graphqlbackend.OptionalResolver{GitHubAppsResolver: NewResolver(logger, db)})
	require.NoError(t, err)

	graphqlbackend.RunTests(t, []*graphqlbackend.Test{{
		Schema:  schema,
		Context: adminCtx,
		Query: `
			mutation CreateGitHubApp($input: CreateGitHubAppInput!) {
				createGitHubApp(input: $input)
			}`,
		ExpectedResult: `{
			"createGitHubApp": 1
		}`,
		Variables: map[string]any{
			"input": map[string]any{
				"appID":        123,
				"baseURL":      "https://github.com",
				"clientID":     "client-id",
				"clientSecret": "client-secret",
				"privateKey":   "private-key",
			},
		},
	}, {
		Schema:  schema,
		Context: userCtx,
		Query: `
			mutation CreateGitHubApp($input: CreateGitHubAppInput!) {
				createGitHubApp(input: $input)
			}`,
		ExpectedResult: `{"createGitHubApp":null}`,
		ExpectedErrors: []*gqlerrors.QueryError{{
			Message: "must be site admin",
			Path:    []any{string("createGitHubApp")},
		}},
		Variables: map[string]any{
			"input": map[string]any{
				"appID":        123,
				"baseURL":      "https://github.com",
				"clientID":     "client-id",
				"clientSecret": "client-secret",
				"privateKey":   "private-key",
			},
		},
	}})
}
