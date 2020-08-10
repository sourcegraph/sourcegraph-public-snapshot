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

func TestExternalServices(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		t.Run("read someone else's external services", func(t *testing.T) {
			db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
				return &types.User{ID: 1}, nil
			}
			db.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
				return &types.User{ID: id}, nil
			}
			defer func() {
				db.Mocks.Users = db.MockUsers{}
			}()

			id := MarshalUserID(2)
			result, err := (&schemaResolver{}).ExternalServices(context.Background(), &ExternalServicesArgs{
				Namespace: &id,
			})
			if want := errMustBeSiteAdminOrSameUser; err != want {
				t.Errorf("err: want %q but got %v", want, err)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})
	})

	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	db.Mocks.ExternalServices.List = func(opt db.ExternalServicesListOptions) ([]*types.ExternalService, error) {
		if opt.NamespaceUserID > 0 {
			return []*types.ExternalService{
				{ID: 1},
			}, nil
		}

		if opt.AfterID > 0 {
			return []*types.ExternalService{
				{ID: 2},
			}, nil
		}

		ess := []*types.ExternalService{
			{ID: 1},
			{ID: 2},
		}
		if opt.LimitOffset != nil {
			return ess[:opt.LimitOffset.Limit], nil
		}
		return ess, nil
	}
	defer func() {
		db.Mocks.Users = db.MockUsers{}
		db.Mocks.ExternalServices = db.MockExternalServices{}
	}()

	gqltesting.RunTests(t, []*gqltesting.Test{
		// Read all external services
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
			{
				externalServices() {
					nodes {
						id
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externalServices": {
					"nodes": [{"id":"RXh0ZXJuYWxTZXJ2aWNlOjE="}, {"id":"RXh0ZXJuYWxTZXJ2aWNlOjI="}]
				}
			}
		`,
		},
		// Read someone's external services
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
			{
				externalServices(namespace: "VXNlcjoy") {
					nodes {
						id
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externalServices": {
					"nodes": [{"id":"RXh0ZXJuYWxTZXJ2aWNlOjE="}]
				}
			}
		`,
		},
		// Pagination
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
			{
				externalServices(first: 1) {
					nodes {
						id
					}
					pageInfo {
						endCursor
						hasNextPage
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externalServices": {
					"nodes":[{"id":"RXh0ZXJuYWxTZXJ2aWNlOjE="}],
					"pageInfo":{"endCursor":"RXh0ZXJuYWxTZXJ2aWNlOjE=","hasNextPage":true}
				}
			}
		`,
		},
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
			{
				externalServices(after: "RXh0ZXJuYWxTZXJ2aWNlOjE=") {
					nodes {
						id
					}
					pageInfo {
						endCursor
						hasNextPage
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externalServices": {
					"nodes":[{"id":"RXh0ZXJuYWxTZXJ2aWNlOjI="}],
					"pageInfo":{"endCursor":null,"hasNextPage":false}
				}
			}
		`,
		},
	})
}
