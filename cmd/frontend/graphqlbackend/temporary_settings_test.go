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
	db := database.NewDB(nil)

	RunTests(t, []*Test{
		{
			// No actor set on context.
			Context: context.Background(),
			Schema:  mustParseGraphQLSchema(t, db),
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
	db := database.NewDB(nil)

	RunTests(t, []*Test{
		{
			Context: actor.WithActor(context.Background(), actor.FromUser(1)),
			Schema:  mustParseGraphQLSchema(t, db),
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

	calledOverwriteTemporarySettings := false
	database.Mocks.TemporarySettings.OverwriteTemporarySettings = func(ctx context.Context, userID int32, contents string) error {
		calledOverwriteTemporarySettings = true
		return nil
	}

	wantErr := errors.New("not authenticated")
	db := database.NewDB(nil)

	RunTests(t, []*Test{
		{
			// No actor set on context.
			Context: context.Background(),
			Schema:  mustParseGraphQLSchema(t, db),
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

	if calledOverwriteTemporarySettings {
		t.Fatal("should not call OverwriteTemporarySettings")
	}
}

func TestOverwriteTemporarySettings(t *testing.T) {
	resetMocks()

	calledOverwriteTemporarySettings := false
	var calledOverwriteTemporarySettingsUserID int32
	database.Mocks.TemporarySettings.OverwriteTemporarySettings = func(ctx context.Context, userID int32, contents string) error {
		calledOverwriteTemporarySettingsUserID = userID
		calledOverwriteTemporarySettings = true
		return nil
	}
	db := database.NewDB(nil)

	RunTests(t, []*Test{
		{
			Context: actor.WithActor(context.Background(), actor.FromUser(1)),
			Schema:  mustParseGraphQLSchema(t, db),
			Query: `
				mutation OverwriteTemporarySettings {
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

	if !calledOverwriteTemporarySettings {
		t.Fatal("should call OverwriteTemporarySettings")
	}
	if calledOverwriteTemporarySettingsUserID != 1 {
		t.Fatalf("should call OverwriteTemporarySettings with userID=1, got=%d", calledOverwriteTemporarySettingsUserID)
	}
}

func TestEditTemporarySettings(t *testing.T) {
	resetMocks()

	calledEditTemporarySettings := false
	var calledEditTemporarySettingsUserID int32
	database.Mocks.TemporarySettings.EditTemporarySettings = func(ctx context.Context, userID int32, settingsToEdit string) error {
		calledEditTemporarySettingsUserID = userID
		calledEditTemporarySettings = true
		return nil
	}
	db := database.NewDB(nil)

	RunTests(t, []*Test{
		{
			Context: actor.WithActor(context.Background(), actor.FromUser(1)),
			Schema:  mustParseGraphQLSchema(t, db),
			Query: `
				mutation EditTemporarySettings {
					editTemporarySettings(
						settingsToEdit: "{\"search.collapsedSidebarSections\": []}"
					) {
						alwaysNil
					}
				}
			`,
			ExpectedResult: "{\"editTemporarySettings\":{\"alwaysNil\":null}}",
		},
	})

	if !calledEditTemporarySettings {
		t.Fatal("should call EditTemporarySettings")
	}
	if calledEditTemporarySettingsUserID != 1 {
		t.Fatalf("should call EditTemporarySettings with userID=1, got=%d", calledEditTemporarySettingsUserID)
	}
}
