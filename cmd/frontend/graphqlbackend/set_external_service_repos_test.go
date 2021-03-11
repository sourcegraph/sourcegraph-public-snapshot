package graphqlbackend

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSetExternalServiceRepos(t *testing.T) {
	database.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
		return &types.ExternalService{
			DisplayName:     "test",
			NamespaceUserID: 1,
			Kind:            extsvc.KindGitHub,
			Config: `{
				  "authorization": {},
				  "repositoryQuery": [
				  ],
				  "token": "not_actually_a_real_token_that_would_be_silly",
				  "url": "https://github.com"
			}`,
		}, nil
	}
	database.Mocks.Users.GetByID = func(ctx context.Context, userID int32) (*types.User, error) {
		return &types.User{
			ID:        userID,
			SiteAdmin: userID == 1,
		}, nil
	}
	database.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1, SiteAdmin: true}, nil
	}
	var called bool
	database.Mocks.ExternalServices.Upsert = func(ctx context.Context, services ...*types.ExternalService) error {
		called = true
		if len(services) != 1 {
			return fmt.Errorf("Expected 1, got %v", len(services))
		}
		svc := services[0]
		cfg, err := svc.Configuration()
		if err != nil {
			return fmt.Errorf("Expected nil, got %s", err)
		}
		gh, ok := cfg.(*schema.GitHubConnection)
		if !ok {
			return fmt.Errorf("Expected *schema.GitHubConnection, got %T", cfg)
		}
		if expected, got := []string{"foo", "bar", "baz"}, gh.Repos; !reflect.DeepEqual(expected, got) {
			return fmt.Errorf("Expected %s got %s", expected, got)
		}
		return nil
	}
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{
		Internal: true,
		UID:      1,
	})

	repoupdater.DefaultClient.HTTPClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
			}, nil
		}),
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: ctx,
			Schema:  mustParseGraphQLSchema(t),
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
	if !called {
		t.Errorf("expected upsert to have been called, but it wasn't")
	}
}
