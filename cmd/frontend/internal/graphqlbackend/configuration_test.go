package graphqlbackend

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMergeConfigs(t *testing.T) {
	orig := deeplyMergedConfigFields
	deeplyMergedConfigFields = map[string]struct{}{"testDeeplyMergedField": struct{}{}}
	defer func() { deeplyMergedConfigFields = orig }()

	tests := map[string]struct {
		configs []string
		want    string
	}{
		"empty": {
			configs: []string{},
			want:    `{}`,
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
				t.Fatal(err)
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
