package graphqlbackend

import (
	"encoding/json"
	"reflect"
	"testing"
)

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
		"deeply merged fields of arrays": {
			configs: []string{
				`{"testDeeplyMergedField":[0,1]}`,
				`{"testDeeplyMergedField":[2,3]}`,
			},
			want: `{"testDeeplyMergedField":[0,1,2,3]}`,
		},
		"deeply merged fields of objects": {
			configs: []string{
				`{"testDeeplyMergedField":{"a":1,"b":2}}`,
				`{"testDeeplyMergedField":{"a":3,"c":4}}`,
			},
			want: `{"testDeeplyMergedField":{"a":3,"b":2,"c":4}}`,
		},
		"deeply merged fields of nested objects": {
			configs: []string{
				`{"testDeeplyMergedField":{"a":{"x":1,"y":2}}}`,
				`{"testDeeplyMergedField":{"a":{"x":3,"z":4}}}`,
			},
			// NOTE: It is expected that this does not include the "y":2 property. Recursive merging
			// may be desirable in the future, but it is not implemented now.
			want: `{"testDeeplyMergedField":{"a":{"x":3,"z":4}}}`,
		},
		"deeply merged fields of arrays and null": {
			configs: []string{
				`{"testDeeplyMergedField":[0,1]}`,
				`{"testDeeplyMergedField":null}`,
				`{"testDeeplyMergedField":[2,3]}`,
			},
			want: `{"testDeeplyMergedField":[0,1,2,3]}`,
		},
		"deeply merged fields with unset 1nd": {
			configs: []string{
				`{}`,
				`{"testDeeplyMergedField":[0,1]}`,
			},
			want: `{"testDeeplyMergedField":[0,1]}`,
		},
		"deeply merged fields with unset 2nd": {
			configs: []string{
				`{"testDeeplyMergedField":[0,1]}`,
				`{}`,
			},
			want: `{"testDeeplyMergedField":[0,1]}`,
		},
		"deeply merged fields of arrays of heterogenous objects": {
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
