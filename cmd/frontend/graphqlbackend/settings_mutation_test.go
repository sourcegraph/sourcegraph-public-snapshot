package graphqlbackend

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSettingsMutation_EditSettings(t *testing.T) {
	users := dbmock.NewMockUserStore()
	users.GetByIDFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: false}, nil)

	settings := dbmock.NewMockSettingsStore()
	settings.GetLatestFunc.SetDefaultReturn(&api.Settings{ID: 1, Contents: "{}"}, nil)
	settings.CreateIfUpToDateFunc.SetDefaultHook(func(ctx context.Context, subject api.SettingsSubject, lastID, authorUserID *int32, contents string) (*api.Settings, error) {
		if want := `{
  "p": {
    "x": 123
  }
}`; contents != want {
			t.Errorf("got %q, want %q", contents, want)
		}
		return &api.Settings{ID: 2, Contents: contents}, nil
	})

	db := dbmock.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.SettingsFunc.SetDefaultReturn(settings)

	RunTests(t, []*Test{
		{
			Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
			Schema:  mustParseGraphQLSchema(t, db),
			Query: `
				mutation($value: JSONValue) {
					settingsMutation(input: {subject: "VXNlcjox", lastID: 1}) {
						editSettings(edit: {keyPath: [{property: "p"}], value: $value}) {
							empty {
								alwaysNil
							}
						}
					}
				}
			`,
			Variables: map[string]interface{}{"value": map[string]int{"x": 123}},
			ExpectedResult: `
				{
					"settingsMutation": {
						"editSettings": {
							"empty": null
						}
					}
				}
			`,
		},
	})
}

func TestSettingsMutation_OverwriteSettings(t *testing.T) {
	users := dbmock.NewMockUserStore()
	users.GetByIDFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: false}, nil)

	settings := dbmock.NewMockSettingsStore()
	settings.GetLatestFunc.SetDefaultReturn(&api.Settings{ID: 1, Contents: "{}"}, nil)
	settings.CreateIfUpToDateFunc.SetDefaultHook(func(ctx context.Context, subject api.SettingsSubject, lastID, authorUserID *int32, contents string) (*api.Settings, error) {
		if want := `x`; contents != want {
			t.Errorf("got %q, want %q", contents, want)
		}
		return &api.Settings{ID: 2, Contents: contents}, nil
	})

	db := dbmock.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.SettingsFunc.SetDefaultReturn(settings)

	RunTests(t, []*Test{
		{
			Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
			Schema:  mustParseGraphQLSchema(t, db),
			Query: `
				mutation($contents: String!) {
					settingsMutation(input: {subject: "VXNlcjox", lastID: 1}) {
						overwriteSettings(contents: $contents) {
							empty {
								alwaysNil
							}
						}
					}
				}
			`,
			Variables: map[string]interface{}{"contents": "x"},
			ExpectedResult: `
				{
					"settingsMutation": {
						"overwriteSettings": {
							"empty": null
						}
					}
				}
			`,
		},
	})
}
