package graphqlbackend

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/google/go-cmp/cmp"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAddExternalService(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
		users.GetByIDFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
		users.TagsFunc.SetDefaultReturn(map[string]bool{}, nil)

		t.Run("user mode not enabled and no namespace", func(t *testing.T) {
			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			result, err := newSchemaResolver(db).AddExternalService(ctx, &addExternalServiceArgs{})
			if want := backend.ErrMustBeSiteAdmin; err != want {
				t.Errorf("err: want %q but got %q", want, err)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("user mode not enabled and has namespace", func(t *testing.T) {
			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			userID := MarshalUserID(1)
			result, err := newSchemaResolver(db).AddExternalService(ctx, &addExternalServiceArgs{
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

			users := database.NewMockUserStoreFrom(users)
			users.CurrentUserAllowedExternalServicesFunc.SetDefaultReturn(conf.ExternalServiceModePublic, nil)

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			userID := MarshalUserID(2)
			result, err := newSchemaResolver(db).AddExternalService(ctx, &addExternalServiceArgs{
				Input: addExternalServiceInput{
					Namespace: &userID,
				},
			})

			want := "the namespace is not the same as the authenticated user"
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

			externalServices := database.NewMockExternalServiceStore()
			externalServices.CreateFunc.SetDefaultReturn(nil)

			users := database.NewMockUserStoreFrom(users)
			users.CurrentUserAllowedExternalServicesFunc.SetDefaultReturn(conf.ExternalServiceModePublic, nil)

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			userID := int32(1)
			gqlID := MarshalUserID(userID)

			result, err := newSchemaResolver(db).AddExternalService(ctx, &addExternalServiceArgs{
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

			externalServices := database.NewMockExternalServiceStore()
			externalServices.CreateFunc.SetDefaultReturn(nil)

			users := database.NewMockUserStoreFrom(users)
			users.CurrentUserAllowedExternalServicesFunc.SetDefaultReturn(conf.ExternalServiceModePublic, nil)
			users.GetByIDFunc.SetDefaultReturn(
				&types.User{ID: 1, Tags: []string{database.TagAllowUserExternalServicePublic}},
				nil,
			)

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			userID := int32(1)
			gqlID := MarshalUserID(userID)

			result, err := newSchemaResolver(db).AddExternalService(ctx, &addExternalServiceArgs{
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

		t.Run("org namespace requested, but feature is not allowed", func(t *testing.T) {
			featureFlags := database.NewMockFeatureFlagStore()
			featureFlags.GetOrgFeatureFlagFunc.SetDefaultReturn(false, nil)

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

			ctx := context.Background()
			orgID := MarshalOrgID(1)
			result, err := newSchemaResolver(db).AddExternalService(ctx, &addExternalServiceArgs{
				Input: addExternalServiceInput{
					Namespace: &orgID,
				},
			})

			want := "organization code host connections are not enabled"
			got := fmt.Sprintf("%v", err)
			if got != want {
				t.Errorf("err: want %q but got %q", want, got)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("org namespace requested, but user does not belong to the org", func(t *testing.T) {
			users := database.NewMockUserStoreFrom(users)
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

			orgMembers := database.NewMockOrgMemberStore()
			orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(nil, nil)

			featureFlags := database.NewMockFeatureFlagStore()
			featureFlags.GetOrgFeatureFlagFunc.SetDefaultReturn(true, nil)

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.OrgMembersFunc.SetDefaultReturn(orgMembers)
			db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			orgID := MarshalOrgID(1)
			result, err := newSchemaResolver(db).AddExternalService(ctx, &addExternalServiceArgs{
				Input: addExternalServiceInput{
					Namespace: &orgID,
				},
			})

			want := "the authenticated user does not belong to the organization requested"
			got := fmt.Sprintf("%v", err)
			if got != want {
				t.Errorf("err: want %q but got %q", want, got)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("org namespace requested, and user belongs to the same org", func(t *testing.T) {
			users := database.NewMockUserStoreFrom(users)
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 10, SiteAdmin: true}, nil)

			orgMembers := database.NewMockOrgMemberStore()
			orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(
				&types.OrgMembership{ID: 1, OrgID: 42, UserID: 10},
				nil,
			)

			externalServices := database.NewMockExternalServiceStore()
			externalServices.CreateFunc.SetDefaultReturn(nil)

			featureFlags := database.NewMockFeatureFlagStore()
			featureFlags.GetOrgFeatureFlagFunc.SetDefaultReturn(true, nil)

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.OrgMembersFunc.SetDefaultReturn(orgMembers)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)
			db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 10})
			orgID := MarshalOrgID(42)

			result, err := newSchemaResolver(db).AddExternalService(ctx, &addExternalServiceArgs{
				Input: addExternalServiceInput{
					Namespace: &orgID,
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			// We want to check the namespace field is populated
			if result.externalService.NamespaceOrgID != 42 {
				t.Fatal("NamespaceOrgID: want 42 but got #{result.externalService.NamespaceOrgID}")
			}
		})
	})

	t.Run("cloud mode, org namespace requested", func(t *testing.T) {
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(false)

		userID := int32(3)
		orgID := int32(45)
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: userID, SiteAdmin: true}, nil)

		orgMembers := database.NewMockOrgMemberStore()
		orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(
			&types.OrgMembership{ID: 2, OrgID: orgID, UserID: 3},
			nil,
		)

		featureFlags := database.NewMockFeatureFlagStore()
		featureFlags.GetOrgFeatureFlagFunc.SetDefaultReturn(true, nil)

		externalServices := database.NewMockExternalServiceStore()
		externalServices.CreateFunc.SetDefaultReturn(nil)

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.OrgMembersFunc.SetDefaultReturn(orgMembers)
		db.ExternalServicesFunc.SetDefaultReturn(externalServices)
		db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: userID})
		marshaledOrgID := MarshalOrgID(orgID)

		t.Run("service kind is not supported", func(t *testing.T) {
			result, err := newSchemaResolver(db).AddExternalService(ctx, &addExternalServiceArgs{
				Input: addExternalServiceInput{
					Namespace:   &marshaledOrgID,
					Kind:        extsvc.KindBitbucketCloud,
					DisplayName: "Bitbucket",
				},
			})

			if want := backend.ErrExternalServiceKindNotSupported; err != want {
				t.Errorf("got err %v, want %v", err, want)
			}
			if result != nil {
				t.Errorf("got result %v, want nil", result)
			}
		})

		t.Run("org can still add external services", func(t *testing.T) {
			svcs := []*types.ExternalService{
				{
					Kind:           extsvc.KindGitHub,
					NamespaceOrgID: orgID,
					Config:         extsvc.NewEmptyConfig(),
				},
			}

			externalServices := database.NewMockExternalServiceStore()
			externalServices.ListFunc.SetDefaultReturn(svcs, nil)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			_, err := newSchemaResolver(db).AddExternalService(ctx, &addExternalServiceArgs{
				Input: addExternalServiceInput{
					Namespace:   &marshaledOrgID,
					Kind:        extsvc.KindGitLab,
					DisplayName: "GitLab",
					Config:      "{\n  \"url\": \"https://gitlab.com\",\n  \"token\": \"dfdf\",\n  \"projectQuery\": [\n    \"projects?membership=true&archived=no\"\n  ]\n}",
				},
			})
			if err != nil {
				t.Fatal(err)
			}
		})

		t.Run("org has reached the limit for a given kind of service", func(t *testing.T) {
			svcs := []*types.ExternalService{
				{
					Kind:           extsvc.KindGitLab,
					NamespaceOrgID: orgID,
					Config:         extsvc.NewEmptyConfig(),
				},
			}

			externalServices := database.NewMockExternalServiceStore()
			externalServices.ListFunc.SetDefaultReturn(svcs, nil)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			result, err := newSchemaResolver(db).AddExternalService(ctx, &addExternalServiceArgs{
				Input: addExternalServiceInput{
					Namespace:   &marshaledOrgID,
					Kind:        extsvc.KindGitLab,
					DisplayName: "GitLab",
					Config:      "{\n  \"url\": \"https://gitlab.com\",\n  \"token\": \"dfdf\",\n  \"projectQuery\": [\n    \"projects?membership=true&archived=no\"\n  ]\n}",
				},
			})
			if want := backend.ErrExternalServiceLimitPerKindReached; err != want {
				t.Errorf("got err %v, want %v", err, want)
			}
			if result != nil {
				t.Errorf("got result %v, want nil", result)
			}
		})
	})

	t.Run("cloud mode, user mode enabled, user namespace requested", func(t *testing.T) {
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(false)

		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExternalServiceUserMode: "public",
			},
		})
		defer conf.Mock(nil)

		userID := int32(4)
		orgID := int32(46)
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: userID, SiteAdmin: true}, nil)
		users.CurrentUserAllowedExternalServicesFunc.SetDefaultReturn(conf.ExternalServiceModePublic, nil)

		orgMembers := database.NewMockOrgMemberStore()
		orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(
			&types.OrgMembership{ID: 2, OrgID: orgID, UserID: 3},
			nil,
		)

		featureFlags := database.NewMockFeatureFlagStore()
		featureFlags.GetOrgFeatureFlagFunc.SetDefaultReturn(true, nil)

		externalServices := database.NewMockExternalServiceStore()
		externalServices.CreateFunc.SetDefaultReturn(nil)

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.OrgMembersFunc.SetDefaultReturn(orgMembers)
		db.ExternalServicesFunc.SetDefaultReturn(externalServices)
		db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: userID})
		marshaledUserID := MarshalUserID(userID)

		t.Run("service kind is not supported", func(t *testing.T) {
			result, err := newSchemaResolver(db).AddExternalService(ctx, &addExternalServiceArgs{
				Input: addExternalServiceInput{
					Namespace:   &marshaledUserID,
					Kind:        extsvc.KindBitbucketCloud,
					DisplayName: "Bitbucket",
				},
			})

			if want := backend.ErrExternalServiceKindNotSupported; err != want {
				t.Errorf("got err %v, want %v", err, want)
			}
			if result != nil {
				t.Errorf("got result %v, want nil", result)
			}
		})

		t.Run("user can still add external services", func(t *testing.T) {
			svcs := []*types.ExternalService{
				{
					Kind:           extsvc.KindGitHub,
					NamespaceOrgID: orgID,
					Config:         extsvc.NewEmptyConfig(),
				},
			}

			externalServices := database.NewMockExternalServiceStore()
			externalServices.ListFunc.SetDefaultReturn(svcs, nil)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			_, err := newSchemaResolver(db).AddExternalService(ctx, &addExternalServiceArgs{
				Input: addExternalServiceInput{
					Namespace:   &marshaledUserID,
					Kind:        extsvc.KindGitLab,
					DisplayName: "GitLab",
					Config:      "{\n  \"url\": \"https://gitlab.com\",\n  \"token\": \"dfdf\",\n  \"projectQuery\": [\n    \"projects?membership=true&archived=no\"\n  ]\n}",
				},
			})
			if err != nil {
				t.Fatal(err)
			}
		})

		t.Run("user has reached the limit for a given kind of service", func(t *testing.T) {
			svcs := []*types.ExternalService{
				{
					Kind:           extsvc.KindGitLab,
					NamespaceOrgID: orgID,
					Config:         extsvc.NewEmptyConfig(),
				},
			}

			externalServices := database.NewMockExternalServiceStore()
			externalServices.ListFunc.SetDefaultReturn(svcs, nil)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			result, err := newSchemaResolver(db).AddExternalService(ctx, &addExternalServiceArgs{
				Input: addExternalServiceInput{
					Namespace:   &marshaledUserID,
					Kind:        extsvc.KindGitLab,
					DisplayName: "GitLab",
					Config:      "{\n  \"url\": \"https://gitlab.com\",\n  \"token\": \"dfdf\",\n  \"projectQuery\": [\n    \"projects?membership=true&archived=no\"\n  ]\n}",
				},
			})
			if want := backend.ErrExternalServiceLimitPerKindReached; err != want {
				t.Errorf("got err %v, want %v", err, want)
			}
			if result != nil {
				t.Errorf("got result %v, want nil", result)
			}
		})
	})

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	externalServices := database.NewMockExternalServiceStore()
	externalServices.CreateFunc.SetDefaultReturn(nil)

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
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
					namespace { id }
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
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

		t.Run("no namespace", func(t *testing.T) {
			externalServices := database.NewMockExternalServiceStore()
			externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID:     id,
					Config: extsvc.NewEmptyConfig(),
				}, nil
			})

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			result, err := newSchemaResolver(db).UpdateExternalService(ctx, &updateExternalServiceArgs{
				Input: updateExternalServiceInput{
					ID: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
				},
			})
			if want := backend.ErrNoAccessExternalService; err != want {
				t.Errorf("err: want %q but got %v", want, err)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("has mismatched user namespace", func(t *testing.T) {
			userID := int32(2)
			externalServices := database.NewMockExternalServiceStore()
			externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID:              id,
					NamespaceUserID: userID,
					Config:          extsvc.NewEmptyConfig(),
				}, nil
			})

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			result, err := newSchemaResolver(db).UpdateExternalService(ctx, &updateExternalServiceArgs{
				Input: updateExternalServiceInput{
					ID: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
				},
			})

			want := backend.ErrNoAccessExternalService.Error()
			got := fmt.Sprintf("%v", err)
			if got != want {
				t.Errorf("err: want %q but got %q", want, got)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("has mismatched org namespace", func(t *testing.T) {
			orgID := int32(42)
			externalServices := database.NewMockExternalServiceStore()
			externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID:             id,
					NamespaceOrgID: orgID,
					Config:         extsvc.NewEmptyConfig(),
				}, nil
			})

			orgMembers := database.NewMockOrgMemberStore()
			orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(nil, nil)

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)
			db.OrgMembersFunc.SetDefaultReturn(orgMembers)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			result, err := newSchemaResolver(db).UpdateExternalService(ctx, &updateExternalServiceArgs{
				Input: updateExternalServiceInput{
					ID: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
				},
			})

			want := backend.ErrNoAccessExternalService.Error()
			got := fmt.Sprintf("%v", err)
			if got != want {
				t.Errorf("err: want %q but got %q", want, got)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("has matching user namespace", func(t *testing.T) {
			userID := int32(1)
			externalServices := database.NewMockExternalServiceStore()
			externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID:              id,
					NamespaceUserID: userID,
					Config:          extsvc.NewEmptyConfig(),
				}, nil
			})
			externalServices.UpdateFunc.SetDefaultReturn(nil)

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			_, err := newSchemaResolver(db).UpdateExternalService(ctx, &updateExternalServiceArgs{
				Input: updateExternalServiceInput{
					ID: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			mockrequire.Called(t, externalServices.UpdateFunc)
		})

		t.Run("has matching org namespace", func(t *testing.T) {
			orgID := int32(1)
			externalServices := database.NewMockExternalServiceStore()
			externalServices.UpdateFunc.SetDefaultReturn(nil)
			externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID:             id,
					NamespaceOrgID: orgID,
					Config:         extsvc.NewEmptyConfig(),
				}, nil
			})

			orgMembers := database.NewMockOrgMemberStore()
			orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultHook(func(_ context.Context, orgID, userID int32) (*types.OrgMembership, error) {
				return &types.OrgMembership{
					OrgID:  orgID,
					UserID: 1,
				}, nil
			})

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.OrgMembersFunc.SetDefaultReturn(orgMembers)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			_, err := newSchemaResolver(db).UpdateExternalService(ctx, &updateExternalServiceArgs{
				Input: updateExternalServiceInput{
					ID: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			mockrequire.Called(t, externalServices.UpdateFunc)
		})
	})

	t.Run("empty config", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

		externalServices := database.NewMockExternalServiceStore()
		externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
			return &types.ExternalService{
				ID:     id,
				Config: extsvc.NewEmptyConfig(),
			}, nil
		})

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.ExternalServicesFunc.SetDefaultReturn(externalServices)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := newSchemaResolver(db).UpdateExternalService(ctx, &updateExternalServiceArgs{
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

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	externalServices := database.NewMockExternalServiceStore()
	externalServices.UpdateFunc.SetDefaultHook(func(ctx context.Context, ps []schema.AuthProviders, id int64, update *database.ExternalServiceUpdate) error {
		cachedUpdate = update
		return nil
	})
	externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
		if cachedUpdate == nil {
			return &types.ExternalService{
				ID:              id,
				NamespaceUserID: userID,
				Kind:            extsvc.KindGitHub,
				Config:          extsvc.NewEmptyConfig(),
			}, nil
		}
		return &types.ExternalService{
			ID:              id,
			Kind:            extsvc.KindGitHub,
			DisplayName:     *cachedUpdate.DisplayName,
			Config:          extsvc.NewUnencryptedConfig(*cachedUpdate.Config),
			NamespaceUserID: userID,
		}, nil
	})

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
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
			Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
		},
	})
}

func TestDeleteExternalService(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

		t.Run("no namespace", func(t *testing.T) {
			externalServices := database.NewMockExternalServiceStore()
			externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID:     id,
					Config: extsvc.NewEmptyConfig(),
				}, nil
			})

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			result, err := newSchemaResolver(db).DeleteExternalService(ctx, &deleteExternalServiceArgs{
				ExternalService: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
			})
			if want := backend.ErrNoAccessExternalService; err != want {
				t.Errorf("err: want %q but got %v", want, err)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("has mismatched user namespace", func(t *testing.T) {
			userID := int32(2)
			externalServices := database.NewMockExternalServiceStore()
			externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID:              id,
					NamespaceUserID: userID,
					Config:          extsvc.NewEmptyConfig(),
				}, nil
			})

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			result, err := newSchemaResolver(db).DeleteExternalService(ctx, &deleteExternalServiceArgs{
				ExternalService: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
			})

			want := backend.ErrNoAccessExternalService.Error()
			got := fmt.Sprintf("%v", err)
			if got != want {
				t.Errorf("err: want %q but got %q", want, got)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("has matching user namespace", func(t *testing.T) {
			userID := int32(1)
			externalServices := database.NewMockExternalServiceStore()
			externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID:              id,
					NamespaceUserID: userID,
					Config:          extsvc.NewEmptyConfig(),
				}, nil
			})
			externalServices.DeleteFunc.SetDefaultReturn(nil)

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			_, err := newSchemaResolver(db).DeleteExternalService(ctx, &deleteExternalServiceArgs{
				ExternalService: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
			})
			if err != nil {
				t.Fatal(err)
			}
			mockrequire.Called(t, externalServices.DeleteFunc)
		})

		t.Run("has mismatched org namespace", func(t *testing.T) {
			orgID := int32(2)
			orgMembers := database.NewMockOrgMemberStore()
			orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(nil, nil)

			externalServices := database.NewMockExternalServiceStore()
			externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID:             id,
					NamespaceOrgID: orgID,
					Config:         extsvc.NewEmptyConfig(),
				}, nil
			})

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)
			db.OrgMembersFunc.SetDefaultReturn(orgMembers)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			result, err := newSchemaResolver(db).DeleteExternalService(ctx, &deleteExternalServiceArgs{
				ExternalService: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
			})

			want := backend.ErrNoAccessExternalService.Error()
			got := fmt.Sprintf("%v", err)
			if got != want {
				t.Errorf("err: want %q but got %q", want, got)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("has matching org namespace", func(t *testing.T) {
			orgID := int32(1)
			orgMembers := database.NewMockOrgMemberStore()
			orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultHook(func(_ context.Context, orgID, userID int32) (*types.OrgMembership, error) {
				return &types.OrgMembership{
					OrgID:  orgID,
					UserID: 1,
				}, nil
			})

			externalServices := database.NewMockExternalServiceStore()
			externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
				return &types.ExternalService{
					ID:             id,
					NamespaceOrgID: orgID,
					Config:         extsvc.NewEmptyConfig(),
				}, nil
			})
			externalServices.DeleteFunc.SetDefaultReturn(nil)

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)
			db.OrgMembersFunc.SetDefaultReturn(orgMembers)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			_, err := newSchemaResolver(db).DeleteExternalService(ctx, &deleteExternalServiceArgs{
				ExternalService: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
			})
			if err != nil {
				t.Fatal(err)
			}
			mockrequire.Called(t, externalServices.DeleteFunc)
		})
	})

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	externalServices := database.NewMockExternalServiceStore()
	externalServices.DeleteFunc.SetDefaultReturn(nil)
	externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
		return &types.ExternalService{
			ID:              id,
			NamespaceUserID: 1,
			Config:          extsvc.NewEmptyConfig(),
		}, nil
	})

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
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
			Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
		},
	})
}

func TestExternalServices(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		t.Run("read users external services", func(t *testing.T) {
			users := database.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
			users.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int32) (*types.User, error) {
				return &types.User{ID: id}, nil
			})
			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			id := MarshalUserID(2)
			result, err := newSchemaResolver(db).ExternalServices(context.Background(), &ExternalServicesArgs{
				Namespace: &id,
			})
			if want := backend.ErrNoAccessExternalService; err != want {
				t.Errorf("err: want %q but got %v", want, err)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("read orgs external services", func(t *testing.T) {
			users := database.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

			orgMembers := database.NewMockOrgMemberStore()
			orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(nil, nil)
			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.OrgMembersFunc.SetDefaultReturn(orgMembers)

			id := MarshalOrgID(2)
			result, err := newSchemaResolver(db).ExternalServices(context.Background(), &ExternalServicesArgs{
				Namespace: &id,
			})
			if want := backend.ErrNoAccessExternalService; err != want {
				t.Errorf("err: want %q but got %v", want, err)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("read site-level external services", func(t *testing.T) {
			users := database.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
			users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
				return &types.User{ID: id}, nil
			})

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			result, err := newSchemaResolver(db).ExternalServices(context.Background(), &ExternalServicesArgs{})
			if want := backend.ErrNoAccessExternalService; err != want {
				t.Errorf("err: want %q but got %v", want, err)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})
	})

	t.Run("authenticated as admin", func(t *testing.T) {
		t.Run("read other users external services", func(t *testing.T) {
			users := database.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
			users.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int32) (*types.User, error) {
				return &types.User{ID: id, SiteAdmin: true}, nil
			})

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			id := MarshalUserID(2)
			result, err := newSchemaResolver(db).ExternalServices(context.Background(), &ExternalServicesArgs{
				Namespace: &id,
			})
			if want := backend.ErrNoAccessExternalService; err != want {
				t.Errorf("err: want %q but got %v", want, err)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})

		t.Run("can read site-level external service", func(t *testing.T) {
			users := database.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
			users.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int32) (*types.User, error) {
				return &types.User{ID: id, SiteAdmin: true}, nil
			})

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			id := MarshalUserID(0)
			_, err := newSchemaResolver(db).ExternalServices(context.Background(), &ExternalServicesArgs{
				Namespace: &id,
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	})

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	externalServices := database.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultHook(func(_ context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
		if opt.NamespaceUserID > 0 {
			return []*types.ExternalService{
				{ID: 1, Config: extsvc.NewEmptyConfig()},
			}, nil
		}

		if opt.AfterID > 0 {
			return []*types.ExternalService{
				{ID: 2, Config: extsvc.NewEmptyConfig()},
			}, nil
		}

		ess := []*types.ExternalService{
			{ID: 1, Config: extsvc.NewEmptyConfig()},
			{ID: 2, Config: extsvc.NewEmptyConfig()},
		}
		if opt.LimitOffset != nil {
			return ess[:opt.LimitOffset.Limit], nil
		}
		return ess, nil
	})
	externalServices.CountFunc.SetDefaultHook(func(ctx context.Context, opt database.ExternalServicesListOptions) (int, error) {
		if opt.NamespaceUserID > 0 || opt.AfterID > 0 {
			return 1, nil
		}

		return 2, nil
	})
	externalServices.GetLastSyncErrorFunc.SetDefaultReturn("Oops", nil)

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	// NOTE: all these tests run as site admin
	RunTests(t, []*Test{
		// Read all external services
		{
			Schema: mustParseGraphQLSchema(t, db),
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
		// Not allowed to read someone else's external service
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
			{
				externalServices(namespace: "VXNlcjoy") {
					nodes {
						id
					}
				}
			}
		`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Path:          []any{"externalServices"},
					Message:       backend.ErrNoAccessExternalService.Error(),
					ResolverError: backend.ErrNoAccessExternalService,
				},
			},
			ExpectedResult: `null`,
		},
		// LastSyncError included
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
			{
				externalServices(namespace: "VXNlcjow") {
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
					"nodes": [
                        {"id":"RXh0ZXJuYWxTZXJ2aWNlOjE=","lastSyncError":"Oops"},
                        {"id":"RXh0ZXJuYWxTZXJ2aWNlOjI=","lastSyncError":"Oops"}
                    ]
				}
			}
		`,
		},
		// Pagination
		{
			Schema: mustParseGraphQLSchema(t, db),
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
			Schema: mustParseGraphQLSchema(t, db),
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
	cmpOpts := cmp.AllowUnexported(graphqlutil.PageInfo{})
	tests := []struct {
		name         string
		opt          database.ExternalServicesListOptions
		mockList     func(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error)
		mockCount    func(ctx context.Context, opt database.ExternalServicesListOptions) (int, error)
		wantPageInfo *graphqlutil.PageInfo
	}{
		{
			name: "no limit set",
			mockList: func(_ context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return []*types.ExternalService{{ID: 1, Config: extsvc.NewEmptyConfig()}}, nil
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
			mockList: func(_ context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return []*types.ExternalService{{ID: 1, Config: extsvc.NewEmptyConfig()}}, nil
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
			mockList: func(_ context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return []*types.ExternalService{{ID: 1, Config: extsvc.NewEmptyConfig()}}, nil
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
			mockList: func(_ context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return []*types.ExternalService{{ID: 1, Config: extsvc.NewEmptyConfig()}}, nil
			},
			mockCount: func(ctx context.Context, opt database.ExternalServicesListOptions) (int, error) {
				return 2, nil
			},
			wantPageInfo: graphqlutil.NextPageCursor(string(MarshalExternalServiceID(1))),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			externalServices := database.NewMockExternalServiceStore()
			externalServices.ListFunc.SetDefaultHook(test.mockList)
			externalServices.CountFunc.SetDefaultHook(test.mockCount)

			db := database.NewMockDB()
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

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

func TestSyncExternalService_ContextTimeout(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Since the timeout in our test is set to 0ms, we do not need to sleep at all. If our code
		// is correct, this handler should timeout right away.
		w.WriteHeader(http.StatusOK)
	}))

	t.Cleanup(func() { s.Close() })

	ctx := context.Background()
	svc := &types.ExternalService{
		Config: extsvc.NewEmptyConfig(),
	}

	err := backend.SyncExternalService(ctx, logtest.Scoped(t), svc, 0*time.Millisecond, repoupdater.NewClient(s.URL))

	if err == nil {
		t.Error("Expected error but got nil")
	}

	expected := "context deadline exceeded"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected error: %q, but got %v", expected, err)
	}
}
