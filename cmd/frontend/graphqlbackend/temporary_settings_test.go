package graphqlbackend

import (
	"context"
	"testing"

	"github.com/cockroachdb/errors"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	ts "github.com/sourcegraph/sourcegraph/internal/temporarysettings"
)

func TestTemporarySettingsNotSignedIn(t *testing.T) {
	resetMocks()

	calledGetTemporarySettings := false
	database.Mocks.TemporarySettings.GetTemporarySettings = func(ctx context.Context, userID int32) (*ts.TemporarySettings, error) {
		calledGetTemporarySettings = true
		return &ts.TemporarySettings{Contents: "{\"search.collapsedSidebarSections\": {\"types\": false}}"}, nil
	}

	wantErr := errors.New("not authenticated")

	RunTests(t, []*Test{
		{
			// No actor set on context.
			Context: context.Background(),
			Schema:  mustParseGraphQLSchema(t),
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

	calledGetTemporarySettings := false
	var calledGetTemporarySettingsUserID int32
	database.Mocks.TemporarySettings.GetTemporarySettings = func(ctx context.Context, userID int32) (*ts.TemporarySettings, error) {
		calledGetTemporarySettings = true
		calledGetTemporarySettingsUserID = userID
		return &ts.TemporarySettings{Contents: "{\"search.collapsedSidebarSections\": {\"types\": false}}"}, nil
	}

	RunTests(t, []*Test{
		{
			Context: actor.WithActor(context.Background(), actor.FromUser(1)),
			Schema:  mustParseGraphQLSchema(t),
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
	if calledGetTemporarySettingsUserID != 1 {
		t.Fatalf("should call GetTemporarySettings with userID=1, got=%d", calledGetTemporarySettingsUserID)
	}
}

func TestOverwriteTemporarySettingsNotSignedIn(t *testing.T) {
	resetMocks()

	calledUpsertTemporarySettings := false
	database.Mocks.TemporarySettings.UpsertTemporarySettings = func(ctx context.Context, userID int32, contents string) error {
		calledUpsertTemporarySettings = true
		return nil
	}

	wantErr := errors.New("not authenticated")

	RunTests(t, []*Test{
		{
			// No actor set on context.
			Context: context.Background(),
			Schema:  mustParseGraphQLSchema(t),
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

	calledUpsertTemporarySettings := false
	var calledUpsertTemporarySettingsUserID int32
	database.Mocks.TemporarySettings.UpsertTemporarySettings = func(ctx context.Context, userID int32, contents string) error {
		calledUpsertTemporarySettingsUserID = userID
		calledUpsertTemporarySettings = true
		return nil
	}

	RunTests(t, []*Test{
		{
			Context: actor.WithActor(context.Background(), actor.FromUser(1)),
			Schema:  mustParseGraphQLSchema(t),
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
	if calledUpsertTemporarySettingsUserID != 1 {
		t.Fatalf("should call UpsertTemporarySettings with userID=1, got=%d", calledUpsertTemporarySettingsUserID)
	}
}
