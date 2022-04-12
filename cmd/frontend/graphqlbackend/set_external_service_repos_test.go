package graphqlbackend

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSetExternalServiceRepos(t *testing.T) {
	externalServices := database.NewMockExternalServiceStore()
	externalServices.GetByIDFunc.SetDefaultReturn(
		&types.ExternalService{
			DisplayName:     "test",
			NamespaceUserID: 1,
			Kind:            extsvc.KindGitHub,
			Config: `
{
	"authorization": {},
	"repositoryQuery": [],
	"token": "not_actually_a_real_token_that_would_be_silly",
	"url": "https://github.com"
}`,
		},
		nil,
	)
	externalServices.UpsertFunc.SetDefaultHook(func(ctx context.Context, services ...*types.ExternalService) error {
		require.Len(t, services, 1)

		svc := services[0]
		cfg, err := svc.Configuration()
		require.NoError(t, err)
		require.IsType(t, &schema.GitHubConnection{}, cfg)

		gh := cfg.(*schema.GitHubConnection)
		assert.Equal(t, []string{"foo", "bar", "baz"}, gh.Repos)
		return nil
	})

	users := database.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{
			ID:        id,
			SiteAdmin: id == 1,
		}, nil
	})
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{
		Internal: true,
		UID:      1,
	})

	db := database.NewMockDB()
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.UsersFunc.SetDefaultReturn(users)

	oldClient := repoupdater.DefaultClient.HTTPClient
	repoupdater.DefaultClient.HTTPClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte{})),
			}, nil
		}),
	}
	defer func() {
		repoupdater.DefaultClient.HTTPClient = oldClient
	}()

	RunTests(t, []*Test{
		{
			Context: ctx,
			Schema:  mustParseGraphQLSchema(t, db),
			Query: `
			mutation {
				setExternalServiceRepos(
					id: "RXh0ZXJuYWxTZXJ2aWNlOjIx"
					allRepos: false
					repos: ["foo","bar","baz"]
				) {
					alwaysNil
				}
			}
			`,
			ExpectedResult: `{"setExternalServiceRepos":{"alwaysNil":null}}`,
		},
	})

	mockrequire.Called(t, externalServices.UpsertFunc)
}
