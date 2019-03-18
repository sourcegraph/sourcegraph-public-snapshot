package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestConfigurationMutation_EditConfiguration(t *testing.T) {
	resetMocks()
	db.Mocks.Users.GetByID = func(context.Context, int32) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	db.Mocks.Settings.GetLatest = func(context.Context, api.ConfigurationSubject) (*api.Settings, error) {
		return &api.Settings{ID: 1, Contents: "{}"}, nil
	}
	db.Mocks.Settings.CreateIfUpToDate = func(ctx context.Context, subject api.ConfigurationSubject, lastID *int32, authorUserID int32, contents string) (*api.Settings, error) {
		if want := `{
  "p": {
    "x": 123
  }
}`; contents != want {
			t.Errorf("got %q, want %q", contents, want)
		}
		return &api.Settings{ID: 2, Contents: contents}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
			Schema:  GraphQLSchema,
			Query: `
				mutation($value: JSONValue) {
					configurationMutation(input: {subject: "VXNlcjox", lastID: 1}) {
						editConfiguration(edit: {keyPath: [{property: "p"}], value: $value}) {
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
					"configurationMutation": {
						"editConfiguration": {
							"empty": null
						}
					}
				}
			`,
		},
	})
}

func TestConfigurationMutation_OverwriteConfiguration(t *testing.T) {
	resetMocks()
	db.Mocks.Users.GetByID = func(context.Context, int32) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	db.Mocks.Settings.GetLatest = func(context.Context, api.ConfigurationSubject) (*api.Settings, error) {
		return &api.Settings{ID: 1, Contents: "{}"}, nil
	}
	db.Mocks.Settings.CreateIfUpToDate = func(ctx context.Context, subject api.ConfigurationSubject, lastID *int32, authorUserID int32, contents string) (*api.Settings, error) {
		if want := `x`; contents != want {
			t.Errorf("got %q, want %q", contents, want)
		}
		return &api.Settings{ID: 2, Contents: contents}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
			Schema:  GraphQLSchema,
			Query: `
				mutation($contents: String!) {
					configurationMutation(input: {subject: "VXNlcjox", lastID: 1}) {
						overwriteConfiguration(contents: $contents) {
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
					"configurationMutation": {
						"overwriteConfiguration": {
							"empty": null
						}
					}
				}
			`,
		},
	})
}
