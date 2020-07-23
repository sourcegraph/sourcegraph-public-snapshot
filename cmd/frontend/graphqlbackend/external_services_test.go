package graphqlbackend

import (
	"context"
	"fmt"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAddExternalService(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{}, nil
		}
		t.Cleanup(func() {
			db.Mocks.Users = db.MockUsers{}
		})

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&schemaResolver{}).AddExternalService(ctx, nil)
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
	db.Mocks.ExternalServices.Create = func(ctx context.Context, confGet func() *conf.Unified, externalService *types.ExternalService) error {
		return nil
	}
	t.Cleanup(func() {
		db.Mocks.Users = db.MockUsers{}
		db.Mocks.ExternalServices = db.MockExternalServices{}
	})

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
			mutation {
				addExternalService(input: {
					kind: GITHUB,
					displayName: "GITHUB #1",
					config: "{\"url\": \"https://github.com\", \"repositoryQuery\": [\"none\"], \"token\": \"abc\"}"
				}) {
					kind
					displayName
					config
				}
			}
		`,
			ExpectedResult: `
			{
				"addExternalService": {
				  "kind": "GITHUB",
				  "displayName": "GITHUB #1",
				  "config": "{\"url\": \"https://github.com\", \"repositoryQuery\": [\"none\"], \"token\": \"abc\"}"
				}
			}
		`,
		},
	})
}

func TestUpdateExternalService(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{}, nil
		}
		t.Cleanup(func() {
			db.Mocks.Users = db.MockUsers{}
		})

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&schemaResolver{}).UpdateExternalService(ctx, nil)
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %v", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	t.Run("empty config", func(t *testing.T) {
		db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{SiteAdmin: true}, nil
		}
		t.Cleanup(func() {
			db.Mocks.Users = db.MockUsers{}
		})

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&schemaResolver{}).UpdateExternalService(ctx, &struct {
			Input UpdateExternalServiceInput
		}{
			Input: UpdateExternalServiceInput{
				ID:     "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
				Config: strptr(""),
			},
		})
		gotErr := fmt.Sprintf("%v", err)
		wantErr := "blank external service configuration is invalid (must be valid JSONC)"
		if gotErr != wantErr {
			t.Errorf("err: want %q but got %q", wantErr, gotErr)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	cachedUpdate := &db.ExternalServiceUpdate{}
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	db.Mocks.ExternalServices.Update = func(ctx context.Context, ps []schema.AuthProviders, id int64, update *db.ExternalServiceUpdate) error {
		cachedUpdate = update
		return nil
	}
	db.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
		return &types.ExternalService{
			ID:          id,
			DisplayName: *cachedUpdate.DisplayName,
			Config:      *cachedUpdate.Config,
		}, nil
	}
	t.Cleanup(func() {
		db.Mocks.Users = db.MockUsers{}
		db.Mocks.ExternalServices = db.MockExternalServices{}
	})

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
			mutation {
				updateExternalService(input: {
					id: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
					displayName: "GITHUB #2",
					config: "{\"url\": \"https://github.com\", \"repositoryQuery\": [\"none\"], \"token\": \"def\"}"
				}) {
					displayName
					config
				}
			}
		`,
			ExpectedResult: `
			{
				"updateExternalService": {
				  "displayName": "GITHUB #2",
				  "config": "{\"url\": \"https://github.com\", \"repositoryQuery\": [\"none\"], \"token\": \"def\"}"
				}
			}
		`,
		},
	})
}

func TestDeleteExternalService(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{}, nil
		}
		t.Cleanup(func() {
			db.Mocks.Users = db.MockUsers{}
		})

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&schemaResolver{}).DeleteExternalService(ctx, nil)
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
	db.Mocks.ExternalServices.Delete = func(ctx context.Context, id int64) error {
		return nil
	}
	db.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
		return &types.ExternalService{
			ID: id,
		}, nil
	}
	t.Cleanup(func() {
		db.Mocks.Users = db.MockUsers{}
		db.Mocks.ExternalServices = db.MockExternalServices{}
	})

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
			mutation {
				deleteExternalService(externalService: "RXh0ZXJuYWxTZXJ2aWNlOjQ=") {
					alwaysNil
				}
			}
		`,
			ExpectedResult: `
			{
				"deleteExternalService": {
					"alwaysNil": null
				}
			}
		`,
		},
	})
}
