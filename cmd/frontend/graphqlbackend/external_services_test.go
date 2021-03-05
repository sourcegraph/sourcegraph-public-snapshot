package graphqlbackend

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go/gqltesting"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAddExternalService(t *testing.T) {
	db := new(dbtesting.MockDB)

	t.Run("authenticated as non-admin", func(t *testing.T) {
		database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{ID: 1}, nil
		}
		database.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{ID: 1}, nil
		}
		defer func() {
			database.Mocks.Users = database.MockUsers{}
		}()

		t.Run("user mode not enabled and no namespace", func(t *testing.T) {
			database.Mocks.Users.HasTag = func(ctx context.Context, userID int32, tag string) (bool, error) {
				return false, nil
			}
			defer func() {
				database.Mocks.Users.HasTag = nil
			}()

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			result, err := (&schemaResolver{db: db}).AddExternalService(ctx, &addExternalServiceArgs{})
			if want := backend.ErrMustBeSiteAdmin; err != want {
				t.Errorf("err: want %q but got %q", want, err)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("user mode not enabled and has namespace", func(t *testing.T) {
			database.Mocks.Users.HasTag = func(ctx context.Context, userID int32, tag string) (bool, error) {
				return false, nil
			}
			defer func() {
				database.Mocks.Users.HasTag = nil
			}()
			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			userID := MarshalUserID(1)
			result, err := (&schemaResolver{db: db}).AddExternalService(ctx, &addExternalServiceArgs{
				Input: addExternalServiceInput{
					Namespace: &userID,
				},
			})

			want := "allow users to add external services is not enabled"
			got := fmt.Sprintf("%v", err)
			if got != want {
				t.Errorf("err: want %q but got %q", want, got)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("user mode enabled but has mismatched namespace", func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					ExternalServiceUserMode: "public",
				},
			})
			defer conf.Mock(nil)

			database.Mocks.Users.HasTag = func(ctx context.Context, userID int32, tag string) (bool, error) {
				return false, nil
			}
			defer func() {
				database.Mocks.Users.HasTag = nil
			}()

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			userID := MarshalUserID(2)
			result, err := (&schemaResolver{db: db}).AddExternalService(ctx, &addExternalServiceArgs{
				Input: addExternalServiceInput{
					Namespace: &userID,
				},
			})

			want := "the namespace is not same as the authenticated user"
			got := fmt.Sprintf("%v", err)
			if got != want {
				t.Errorf("err: want %q but got %q", want, got)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("user mode enabled and has matching namespace", func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					ExternalServiceUserMode: "public",
				},
			})
			defer conf.Mock(nil)

			database.Mocks.Users.HasTag = func(ctx context.Context, userID int32, tag string) (bool, error) {
				return false, nil
			}
			defer func() {
				database.Mocks.Users.HasTag = nil
			}()
			database.Mocks.ExternalServices.Create = func(ctx context.Context, confGet func() *conf.Unified, externalService *types.ExternalService) error {
				return nil
			}
			defer func() {
				database.Mocks.ExternalServices = database.MockExternalServices{}
			}()

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			userID := int32(1)
			gqlID := MarshalUserID(userID)
			result, err := (&schemaResolver{db: db}).AddExternalService(ctx, &addExternalServiceArgs{
				Input: addExternalServiceInput{
					Namespace: &gqlID,
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			// We want to check the namespace field is populated
			if result.externalService.NamespaceUserID == 0 {
				t.Fatal("NamespaceUserID: want non-nil but got nil")
			} else if result.externalService.NamespaceUserID != userID {
				t.Fatalf("NamespaceUserID: want %d but got %d", userID, result.externalService.NamespaceUserID)
			}
		})

		t.Run("user mode not enabled but user has public tag", func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					ExternalServiceUserMode: "disabled",
				},
			})
			defer conf.Mock(nil)

			database.Mocks.Users.HasTag = func(ctx context.Context, userID int32, tag string) (bool, error) {
				return true, nil
			}
			defer func() {
				database.Mocks.Users.HasTag = nil
			}()
			database.Mocks.ExternalServices.Create = func(ctx context.Context, confGet func() *conf.Unified, externalService *types.ExternalService) error {
				return nil
			}
			defer func() {
				database.Mocks.ExternalServices = database.MockExternalServices{}
			}()

			database.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
				return &types.User{
					ID: 1,
					Tags: []string{
						database.TagAllowUserExternalServicePublic,
					},
				}, nil
			}
			defer func() {
				database.Mocks.Users = database.MockUsers{}
			}()

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			userID := int32(1)
			gqlID := MarshalUserID(userID)

			result, err := (&schemaResolver{db: db}).AddExternalService(ctx, &addExternalServiceArgs{
				Input: addExternalServiceInput{
					Namespace: &gqlID,
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			// We want to check the namespace field is populated
			if result.externalService.NamespaceUserID == 0 {
				t.Fatal("NamespaceUserID: want non-nil but got nil")
			} else if result.externalService.NamespaceUserID != userID {
				t.Fatalf("NamespaceUserID: want %d but got %d", userID, result.externalService.NamespaceUserID)
			}
		})
	})

	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	database.Mocks.ExternalServices.Create = func(ctx context.Context, confGet func() *conf.Unified, externalService *types.ExternalService) error {
		return nil
	}

	t.Cleanup(func() {
		database.Mocks.Users = database.MockUsers{}
		database.Mocks.ExternalServices = database.MockExternalServices{}
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
					namespace
				}
			}
		`,
			ExpectedResult: `
			{
				"addExternalService": {
					"kind": "GITHUB",
					"displayName": "GITHUB #1",
					"config":"{\n  \"url\": \"https://github.com\",\n  \"repositoryQuery\": [\n    \"none\"\n  ],\n  \"token\": \"` + types.RedactedSecret + `\"\n}",
					"namespace": null
				}
			}
		`,
		},
	})
}

func TestUpdateExternalService(t *testing.T) {
	db := new(dbtesting.MockDB)

	t.Run("authenticated as non-admin", func(t *testing.T) {
		database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{ID: 1}, nil
		}
		defer func() {
			database.Mocks.Users = database.MockUsers{}
		}()

		t.Run("no namespace", func(t *testing.T) {
			database.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID: id,
				}, nil
			}
			defer func() {
				database.Mocks.ExternalServices = database.MockExternalServices{}
			}()

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			result, err := (&schemaResolver{db: db}).UpdateExternalService(ctx, &updateExternalServiceArgs{
				Input: updateExternalServiceInput{
					ID: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
				},
			})
			if want := backend.ErrMustBeSiteAdmin; err != want {
				t.Errorf("err: want %q but got %v", want, err)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("has mismatched namespace", func(t *testing.T) {
			userID := int32(2)
			database.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID:              id,
					NamespaceUserID: userID,
				}, nil
			}
			defer func() {
				database.Mocks.ExternalServices = database.MockExternalServices{}
			}()

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			result, err := (&schemaResolver{db: db}).UpdateExternalService(ctx, &updateExternalServiceArgs{
				Input: updateExternalServiceInput{
					ID: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
				},
			})

			want := "the authenticated user does not have access to this external service"
			got := fmt.Sprintf("%v", err)
			if got != want {
				t.Errorf("err: want %q but got %q", want, got)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("has matching namespace", func(t *testing.T) {
			userID := int32(1)
			database.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID:              id,
					NamespaceUserID: userID,
				}, nil
			}
			calledUpdate := false
			database.Mocks.ExternalServices.Update = func(ctx context.Context, ps []schema.AuthProviders, id int64, update *database.ExternalServiceUpdate) error {
				calledUpdate = true
				return nil
			}
			defer func() {
				database.Mocks.ExternalServices = database.MockExternalServices{}
			}()

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			_, err := (&schemaResolver{db: db}).UpdateExternalService(ctx, &updateExternalServiceArgs{
				Input: updateExternalServiceInput{
					ID: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if !calledUpdate {
				t.Fatal("!calledUpdate")
			}
		})
	})

	t.Run("empty config", func(t *testing.T) {
		database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{SiteAdmin: true}, nil
		}
		database.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
			return &types.ExternalService{
				ID: id,
			}, nil
		}
		defer func() {
			database.Mocks.Users = database.MockUsers{}
			database.Mocks.ExternalServices = database.MockExternalServices{}
		}()

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := (&schemaResolver{db: db}).UpdateExternalService(ctx, &updateExternalServiceArgs{
			Input: updateExternalServiceInput{
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

	userID := int32(1)
	var cachedUpdate *database.ExternalServiceUpdate
	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	database.Mocks.ExternalServices.Update = func(ctx context.Context, ps []schema.AuthProviders, id int64, update *database.ExternalServiceUpdate) error {
		cachedUpdate = update
		return nil
	}
	database.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
		if cachedUpdate == nil {
			return &types.ExternalService{
				ID:              id,
				NamespaceUserID: userID,
				Kind:            extsvc.KindGitHub,
			}, nil
		}
		return &types.ExternalService{
			ID:              id,
			Kind:            extsvc.KindGitHub,
			DisplayName:     *cachedUpdate.DisplayName,
			Config:          *cachedUpdate.Config,
			NamespaceUserID: userID,
		}, nil
	}
	t.Cleanup(func() {
		database.Mocks.Users = database.MockUsers{}
		database.Mocks.ExternalServices = database.MockExternalServices{}
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
				  "config":"{\n  \"url\": \"https://github.com\",\n  \"repositoryQuery\": [\n    \"none\"\n  ],\n  \"token\": \"` + types.RedactedSecret + `\"\n}"

				}
			}
		`,
		},
	})
}

func TestDeleteExternalService(t *testing.T) {
	db := new(dbtesting.MockDB)

	t.Run("authenticated as non-admin", func(t *testing.T) {
		database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{ID: 1}, nil
		}
		defer func() {
			database.Mocks.Users = database.MockUsers{}
		}()

		t.Run("no namespace", func(t *testing.T) {
			database.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID: id,
				}, nil
			}
			defer func() {
				database.Mocks.ExternalServices = database.MockExternalServices{}
			}()

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			result, err := (&schemaResolver{db: db}).DeleteExternalService(ctx, &deleteExternalServiceArgs{
				ExternalService: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
			})
			if want := backend.ErrMustBeSiteAdmin; err != want {
				t.Errorf("err: want %q but got %v", want, err)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("has mismatched namespace", func(t *testing.T) {
			userID := int32(2)
			database.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID:              id,
					NamespaceUserID: userID,
				}, nil
			}
			defer func() {
				database.Mocks.ExternalServices = database.MockExternalServices{}
			}()

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			result, err := (&schemaResolver{db: db}).DeleteExternalService(ctx, &deleteExternalServiceArgs{
				ExternalService: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
			})

			want := "the authenticated user does not have access to this external service"
			got := fmt.Sprintf("%v", err)
			if got != want {
				t.Errorf("err: want %q but got %q", want, got)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("has matching namespace", func(t *testing.T) {
			userID := int32(1)
			database.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID:              id,
					NamespaceUserID: userID,
				}, nil
			}
			calledDelete := false
			database.Mocks.ExternalServices.Delete = func(ctx context.Context, id int64) error {
				calledDelete = true
				return nil
			}
			defer func() {
				database.Mocks.ExternalServices = database.MockExternalServices{}
			}()

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			_, err := (&schemaResolver{db: db}).DeleteExternalService(ctx, &deleteExternalServiceArgs{
				ExternalService: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
			})
			if err != nil {
				t.Fatal(err)
			}
			if !calledDelete {
				t.Fatal("!calledDelete")
			}
		})
	})

	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	database.Mocks.ExternalServices.Delete = func(ctx context.Context, id int64) error {
		return nil
	}
	database.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
		userID := int32(1)
		return &types.ExternalService{
			ID:              id,
			NamespaceUserID: userID,
		}, nil
	}
	t.Cleanup(func() {
		database.Mocks.Users = database.MockUsers{}
		database.Mocks.ExternalServices = database.MockExternalServices{}
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
	db := new(dbtesting.MockDB)

	t.Run("authenticated as non-admin", func(t *testing.T) {
		t.Run("read someone else's external services", func(t *testing.T) {
			database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
				return &types.User{ID: 1}, nil
			}
			database.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
				return &types.User{ID: id}, nil
			}
			defer func() {
				database.Mocks.Users = database.MockUsers{}
			}()

			id := MarshalUserID(2)
			result, err := (&schemaResolver{db: db}).ExternalServices(context.Background(), &ExternalServicesArgs{
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

	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	database.Mocks.ExternalServices.List = func(opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
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
	database.Mocks.ExternalServices.Count = func(ctx context.Context, opt database.ExternalServicesListOptions) (int, error) {
		if opt.NamespaceUserID > 0 || opt.AfterID > 0 {
			return 1, nil
		}

		return 2, nil
	}
	database.Mocks.ExternalServices.GetLastSyncError = func(id int64) (string, error) {
		return "Oops", nil
	}
	defer func() {
		database.Mocks.Users = database.MockUsers{}
		database.Mocks.ExternalServices = database.MockExternalServices{}
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
		// LastSyncError included
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
			{
				externalServices(namespace: "VXNlcjoy") {
					nodes {
						id
						lastSyncError
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externalServices": {
					"nodes": [{"id":"RXh0ZXJuYWxTZXJ2aWNlOjE=","lastSyncError":"Oops"}]
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

func TestExternalServices_PageInfo(t *testing.T) {
	db := new(dbtesting.MockDB)
	cmpOpts := cmp.AllowUnexported(graphqlutil.PageInfo{})
	tests := []struct {
		name         string
		opt          database.ExternalServicesListOptions
		mockList     func(opt database.ExternalServicesListOptions) ([]*types.ExternalService, error)
		mockCount    func(ctx context.Context, opt database.ExternalServicesListOptions) (int, error)
		wantPageInfo *graphqlutil.PageInfo
	}{
		{
			name: "no limit set",
			mockList: func(opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return []*types.ExternalService{{ID: 1}}, nil
			},
			wantPageInfo: graphqlutil.HasNextPage(false),
		},
		{
			name: "less results than the limit",
			opt: database.ExternalServicesListOptions{
				LimitOffset: &database.LimitOffset{
					Limit: 10,
				},
			},
			mockList: func(opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return []*types.ExternalService{{ID: 1}}, nil
			},
			wantPageInfo: graphqlutil.HasNextPage(false),
		},
		{
			name: "same number of results as the limit, and no more",
			opt: database.ExternalServicesListOptions{
				LimitOffset: &database.LimitOffset{
					Limit: 1,
				},
			},
			mockList: func(opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return []*types.ExternalService{{ID: 1}}, nil
			},
			mockCount: func(ctx context.Context, opt database.ExternalServicesListOptions) (int, error) {
				return 1, nil
			},
			wantPageInfo: graphqlutil.HasNextPage(false),
		},
		{
			name: "same number of results as the limit, and has more",
			opt: database.ExternalServicesListOptions{
				LimitOffset: &database.LimitOffset{
					Limit: 1,
				},
			},
			mockList: func(opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return []*types.ExternalService{{ID: 1}}, nil
			},
			mockCount: func(ctx context.Context, opt database.ExternalServicesListOptions) (int, error) {
				return 2, nil
			},
			wantPageInfo: graphqlutil.NextPageCursor(string(marshalExternalServiceID(1))),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			database.Mocks.ExternalServices.List = test.mockList
			database.Mocks.ExternalServices.Count = test.mockCount
			defer func() {
				database.Mocks.ExternalServices = database.MockExternalServices{}
			}()

			r := &externalServiceConnectionResolver{
				db:  db,
				opt: test.opt,
			}
			pageInfo, err := r.PageInfo(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.wantPageInfo, pageInfo, cmpOpts); diff != "" {
				t.Fatalf("PageInfo mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
