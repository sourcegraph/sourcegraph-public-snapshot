package graphqlbackend

import (
	"context"
	"testing"

	"github.com/cockroachdb/errors"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
	ts "github.com/sourcegraph/sourcegraph/internal/temporarysettings"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestTemporarySettingsNotSignedIn(t *testing.T) {
	resetMocks()

	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return nil, database.ErrNoCurrentUser
	}

	calledGetTemporarySettings := false
	database.Mocks.TemporarySettings.GetTemporarySettings = func(ctx context.Context, userID int32) (*ts.TemporarySettings, error) {
		calledGetTemporarySettings = true
		return &ts.TemporarySettings{Contents: "{\"search.collapsedSidebarSections\": {\"types\": false}}"}, nil
	}

	wantErr := errors.New("not authenticated")

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
				query {
					temporarySettings {
						contents
					}
				}
			`,
			ExpectedResult: "null",
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Path:          []interface{}{"temporarySettings"},
					Message:       wantErr.Error(),
					ResolverError: wantErr,
				},
			},
		},
	})

	if calledGetTemporarySettings {
		t.Fatal("should not call GetTemporarySettings")
	}
}

func TestTemporarySettings(t *testing.T) {
	resetMocks()

	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{ID: 1, SiteAdmin: false}, nil
	}

	calledGetTemporarySettings := false
	database.Mocks.TemporarySettings.GetTemporarySettings = func(ctx context.Context, userID int32) (*ts.TemporarySettings, error) {
		calledGetTemporarySettings = true
		return &ts.TemporarySettings{Contents: "{\"search.collapsedSidebarSections\": {\"types\": false}}"}, nil
	}

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
				query {
					temporarySettings {
						contents
					}
				}
			`,
			ExpectedResult: `
				{
					"temporarySettings": {
						"contents": "{\"search.collapsedSidebarSections\": {\"types\": false}}"
					}
				}
			`,
		},
	})

	if !calledGetTemporarySettings {
		t.Fatal("should call GetTemporarySettings")
	}
}

func TestOverwriteTemporarySettingsNotSignedIn(t *testing.T) {
	resetMocks()

	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return nil, database.ErrNoCurrentUser
	}

	calledUpsertTemporarySettings := false
	database.Mocks.TemporarySettings.UpsertTemporarySettings = func(ctx context.Context, userID int32, contents string) error {
		calledUpsertTemporarySettings = true
		return nil
	}

	wantErr := errors.New("not authenticated")

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
				mutation ModifyTemporarySettings {
					overwriteTemporarySettings(
						contents: "{\"search.collapsedSidebarSections\": []}"
					) {
						alwaysNil
					}
				}
			`,
			ExpectedResult: "null",
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Path:          []interface{}{"overwriteTemporarySettings"},
					Message:       wantErr.Error(),
					ResolverError: wantErr,
				},
			},
		},
	})

	if calledUpsertTemporarySettings {
		t.Fatal("should not call UpsertTemporarySettings")
	}
}

func TestOverwriteTemporarySettings(t *testing.T) {
	resetMocks()

	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{ID: 1, SiteAdmin: false}, nil
	}

	calledUpsertTemporarySettings := false
	database.Mocks.TemporarySettings.UpsertTemporarySettings = func(ctx context.Context, userID int32, contents string) error {
		calledUpsertTemporarySettings = true
		return nil
	}

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
				mutation ModifyTemporarySettings {
					overwriteTemporarySettings(
						contents: "{\"search.collapsedSidebarSections\": []}"
					) {
						alwaysNil
					}
				}
			`,
			ExpectedResult: "{\"overwriteTemporarySettings\":{\"alwaysNil\":null}}",
		},
	})

	if !calledUpsertTemporarySettings {
		t.Fatal("should call UpsertTemporarySettings")
	}
}
