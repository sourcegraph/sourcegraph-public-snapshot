package graphqlbackend

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

func TestMergeSettings(t *testing.T) {
	orig := deeplyMergedSettingsFields
	deeplyMergedSettingsFields = map[string]int{
		"f1": 1,
		"f2": 2,
	}
	defer func() { deeplyMergedSettingsFields = orig }()

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
			configs: []string{
				`
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
		"arrays": {
			configs: []string{
				`{"f1":[0,1]}`,
				`{"f1":[2,3]}`,
			},
			want: `{"f1":[0,1,2,3]}`,
		},
		"objects": {
			configs: []string{
				`{"f1":{"a":1,"b":2}}`,
				`{"f1":{"a":3,"c":4}}`,
			},
			want: `{"f1":{"a":3,"b":2,"c":4}}`,
		},
		"nested objects with depth 1": {
			configs: []string{
				`{"f1":{"a":{"x":1,"y":2}}}`,
				`{"f1":{"a":{"x":3,"z":4}}}`,
			},
			// NOTE: It is expected that this does not include the "y":2 property because the
			// merging only occurs 1 level deep for field f1.
			want: `{"f1":{"a":{"x":3,"z":4}}}`,
		},
		"nested objects with depth 2": {
			configs: []string{
				`{"f2":{"a":{"x":1,"y":2}}}`,
				`{"f2":{"a":{"x":3,"z":4}}}`,
			},
			want: `{"f2":{"a":{"x":3,"y":2,"z":4}}}`,
		},
		"arrays and null": {
			configs: []string{
				`{"f1":[0,1]}`,
				`{"f1":null}`,
				`{"f1":[2,3]}`,
			},
			want: `{"f1":[0,1,2,3]}`,
		},
		"unset 1nd": {
			configs: []string{
				`{}`,
				`{"f1":[0,1]}`,
			},
			want: `{"f1":[0,1]}`,
		},
		"unset 2nd": {
			configs: []string{
				`{"f1":[0,1]}`,
				`{}`,
			},
			want: `{"f1":[0,1]}`,
		},
		"arrays of heterogenous objects": {
			configs: []string{
				`{"f1":[{"a":0},1]}`,
				`{"f1":[2,{"b":3}]}`,
			},
			want: `{"f1":[{"a":0},1,2,{"b":3}]}`,
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			merged, err := mergeSettings(test.configs)
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

func TestSubjects(t *testing.T) {
	db := new(dbtesting.MockDB)
	t.Run("Default settings are included", func(t *testing.T) {
		cascade := &settingsCascade{db: db, unauthenticatedActor: true}
		subjects, err := cascade.Subjects(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(subjects) < 1 {
			t.Fatal("Expected at least 1 subject")
		}
		if subjects[0].defaultSettings == nil {
			t.Fatal("Expected the first subject to be default settings")
		}
	})
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
