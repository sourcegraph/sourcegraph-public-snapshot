package graphqlbackend

import (
	"context"
	"testing"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
)

// 🚨 SECURITY: This tests that users can't create tokens for users they aren't allowed to do so for.
func TestMutation_CreateAccessToken(t *testing.T) {
	mockAccessTokensCreate := func(t *testing.T) {
		db.Mocks.AccessTokens.Create = func(userID int32, note string) (int64, string, error) {
			if want := int32(1); userID != want {
				t.Errorf("got %v, want %v", userID, want)
			}
			if want := "n"; note != want {
				t.Errorf("got %q, want %q", note, want)
			}
			return 1, "t", nil
		}
	}

	const uid1GQLID = "VXNlcjox"

	t.Run("authenticated as user", func(t *testing.T) {
		resetMocks()
		mockAccessTokensCreate(t)
		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
				Schema:  GraphQLSchema,
				Query: `
				mutation {
					createAccessToken(user: "` + uid1GQLID + `", note: "n") {
						id
						token
					}
				}
			`,
				ExpectedResult: `
				{
					"createAccessToken": {
						"id": "QWNjZXNzVG9rZW46MQ==",
						"token": "t"
					}
				}
			`,
			},
		})
	})

	t.Run("authenticated as different user who is a site-admin", func(t *testing.T) {
		resetMocks()
		const differentSiteAdminUID = 234
		mockAccessTokensCreate(t)
		db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
			return &types.User{ID: differentSiteAdminUID, SiteAdmin: true}, nil
		}
		defer func() { db.Mocks.Users.GetByCurrentAuthUser = nil }()

		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Context: actor.WithActor(context.Background(), &actor.Actor{UID: differentSiteAdminUID}),
				Schema:  GraphQLSchema,
				Query: `
				mutation {
					createAccessToken(user: "` + uid1GQLID + `", note: "n") {
						id
						token
					}
				}
			`,
				ExpectedResult: `
				{
					"createAccessToken": {
						"id": "QWNjZXNzVG9rZW46MQ==",
						"token": "t"
					}
				}
			`,
			},
		})
	})

	t.Run("unauthenticated", func(t *testing.T) {
		resetMocks()
		db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) { return nil, db.ErrNoCurrentUser }
		defer func() { db.Mocks.Users.GetByCurrentAuthUser = nil }()

		ctx := actor.WithActor(context.Background(), nil)
		result, err := (&schemaResolver{}).CreateAccessToken(ctx, &createAccessTokenInput{User: uid1GQLID, Note: "n"})
		if want := backend.ErrNotAuthenticated; err != want {
			t.Errorf("got err %v, want %v", err, want)
		}
		if result != nil {
			t.Errorf("got result %v, want nil", result)
		}
	})

	t.Run("authenticated as different non-site-admin user", func(t *testing.T) {
		resetMocks()
		const differentNonSiteAdminUID = 456
		db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) { return &types.User{ID: differentNonSiteAdminUID}, nil }
		defer func() { db.Mocks.Users.GetByCurrentAuthUser = nil }()

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: differentNonSiteAdminUID})
		result, err := (&schemaResolver{}).CreateAccessToken(ctx, &createAccessTokenInput{User: uid1GQLID, Note: "n"})
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("got err %v, want %v", err, want)
		}
		if result != nil {
			t.Errorf("got result %v, want nil", result)
		}
	})
}

// 🚨 SECURITY: This tests that users can't delete tokens they shouldn't be allowed to delete.
func TestMutation_DeleteAccessToken(t *testing.T) {
	mockAccessTokens := func(t *testing.T) {
		db.Mocks.AccessTokens.DeleteByID = func(id int64, userID int32) error {
			if want := int64(1); id != want {
				t.Errorf("got %q, want %q", id, want)
			}
			if want := int32(2); userID != want {
				t.Errorf("got %v, want %v", userID, want)
			}
			return nil
		}
		db.Mocks.AccessTokens.GetByID = func(id int64) (*db.AccessToken, error) {
			if want := int64(1); id != want {
				t.Errorf("got %d, want %d", id, want)
			}
			return &db.AccessToken{ID: 1, UserID: 2}, nil
		}
	}

	token1GQLID := graphql.ID("QWNjZXNzVG9rZW46MQ==")

	t.Run("authenticated as user", func(t *testing.T) {
		resetMocks()
		mockAccessTokens(t)
		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Context: actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
				Schema:  GraphQLSchema,
				Query: `
				mutation {
					deleteAccessToken(byID: "` + string(token1GQLID) + `") {
						alwaysNil
					}
				}
			`,
				ExpectedResult: `
				{
					"deleteAccessToken": {
						"alwaysNil": null
					}
				}
			`,
			},
		})
	})

	t.Run("authenticated as different user who is a site-admin", func(t *testing.T) {
		resetMocks()
		const differentSiteAdminUID = 234
		mockAccessTokens(t)
		db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
			return &types.User{ID: differentSiteAdminUID, SiteAdmin: true}, nil
		}
		defer func() { db.Mocks.Users.GetByCurrentAuthUser = nil }()

		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Context: actor.WithActor(context.Background(), &actor.Actor{UID: differentSiteAdminUID}),
				Schema:  GraphQLSchema,
				Query: `
				mutation {
					deleteAccessToken(byID: "` + string(token1GQLID) + `") {
						alwaysNil
					}
				}
			`,
				ExpectedResult: `
				{
					"deleteAccessToken": {
						"alwaysNil": null
					}
				}
			`,
			},
		})
	})

	t.Run("unauthenticated", func(t *testing.T) {
		resetMocks()
		mockAccessTokens(t)
		db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) { return nil, db.ErrNoCurrentUser }
		defer func() { db.Mocks.Users.GetByCurrentAuthUser = nil }()

		ctx := actor.WithActor(context.Background(), nil)
		result, err := (&schemaResolver{}).DeleteAccessToken(ctx, &deleteAccessTokenInput{ByID: &token1GQLID})
		if want := backend.ErrNotAuthenticated; err != want {
			t.Errorf("got err %v, want %v", err, want)
		}
		if result != nil {
			t.Errorf("got result %v, want nil", result)
		}
	})

	t.Run("authenticated as different non-site-admin user", func(t *testing.T) {
		resetMocks()
		const differentNonSiteAdminUID = 456
		mockAccessTokens(t)
		db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) { return &types.User{ID: differentNonSiteAdminUID}, nil }
		defer func() { db.Mocks.Users.GetByCurrentAuthUser = nil }()

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: differentNonSiteAdminUID})
		result, err := (&schemaResolver{}).DeleteAccessToken(ctx, &deleteAccessTokenInput{ByID: &token1GQLID})
		if want := backend.ErrMustBeSiteAdmin; err != want {
			t.Errorf("got err %v, want %v", err, want)
		}
		if result != nil {
			t.Errorf("got result %v, want nil", result)
		}
	})
}
