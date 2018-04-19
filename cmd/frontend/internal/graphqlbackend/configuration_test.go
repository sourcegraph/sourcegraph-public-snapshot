package graphqlbackend

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func TestConfigurationMutation_UpdateConfiguration(t *testing.T) {
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
						updateConfiguration(input: {property: "p", value: $value}) {
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
						"updateConfiguration": {
							"empty": null
						}
					}
				}
			`,
		},
	})
}

func TestMergeConfigs(t *testing.T) {
	orig := deeplyMergedConfigFields
	deeplyMergedConfigFields = map[string]struct{}{"testDeeplyMergedField": {}}
	defer func() { deeplyMergedConfigFields = orig }()

	tests := map[string]struct {
		configs []string
		want    string
		wantErr bool
	}{
		"empty": {
			configs: []string{},
			want:    `{}`,
		},
		"syntax error": {
			configs: []string{`error!`, `{"a":1}`},
			wantErr: true,
		},
		"single": {
			configs: []string{`{"a":1}`},
			want:    `{"a":1}`,
		},
		"single with comments": {
			configs: []string{`
/* comment */
{
	// comment
	"a": 1 // comment
}`,
			},
			want: `{"a":1}`,
		},
		"multiple with no deeply merged fields": {
			configs: []string{
				`{"a":1}`,
				`{"b":2}`,
			},
			want: `{"a":1,"b":2}`,
		},
		"deeply merged fields of strings": {
			configs: []string{
				`{"testDeeplyMergedField":[0,1]}`,
				`{"testDeeplyMergedField":[2,3]}`,
			},
			want: `{"testDeeplyMergedField":[0,1,2,3]}`,
		},
		"deeply merged fields of strings with null": {
			configs: []string{
				`{"testDeeplyMergedField":[0,1]}`,
				`{"testDeeplyMergedField":null}`,
				`{"testDeeplyMergedField":[2,3]}`,
			},
			want: `{"testDeeplyMergedField":[0,1,2,3]}`,
		},
		"deeply merged fields of strings with unset 1nd": {
			configs: []string{
				`{}`,
				`{"testDeeplyMergedField":[0,1]}`,
			},
			want: `{"testDeeplyMergedField":[0,1]}`,
		},
		"deeply merged fields of strings with unset 2nd": {
			configs: []string{
				`{"testDeeplyMergedField":[0,1]}`,
				`{}`,
			},
			want: `{"testDeeplyMergedField":[0,1]}`,
		},
		"deeply merged fields of heterogenous objects": {
			configs: []string{
				`{"testDeeplyMergedField":[{"a":0},1]}`,
				`{"testDeeplyMergedField":[2,{"b":3}]}`,
			},
			want: `{"testDeeplyMergedField":[{"a":0},1,2,{"b":3}]}`,
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			merged, err := mergeConfigs(test.configs)
			if err != nil {
				if test.wantErr {
					return
				}
				t.Fatal(err)
			}
			if test.wantErr {
				t.Fatal("got no error, want error")
			}
			if !jsonDeepEqual(string(merged), test.want) {
				t.Errorf("got %s, want %s", merged, test.want)
			}
		})
	}
}

func jsonDeepEqual(a, b string) bool {
	var va, vb interface{}
	if err := json.Unmarshal([]byte(a), &va); err != nil {
		panic(err)
	}
	if err := json.Unmarshal([]byte(b), &vb); err != nil {
		panic(err)
	}
	return reflect.DeepEqual(va, vb)
}
