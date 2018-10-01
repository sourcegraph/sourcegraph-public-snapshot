package conf

import (
	"reflect"
	"testing"
)

func TestExpand(t *testing.T) {
	vars := map[string]string{
		"v": "1",
	}
	mapping := func(name string) string { return vars[name] }

	tests := map[string]struct {
		interpolatedData string
		seenVars         []string
	}{
		"":                                    {"", nil},
		`"a${v}c"`:                            {`"a1c"`, []string{"v"}},
		`{"a${v}c": "a${v}c"}`:                {`{"a${v}c":"a1c"}`, []string{"v"}},
		`{"a": ["a${v}c"]}`:                   {`{"a":["a1c"]}`, []string{"v"}},
		`{"a": [true, "a${v}c", "$v${v}$v"]}`: {`{"a":[true,"a1c","111"]}`, []string{"v"}},
		`{"a": [true, {"b": "a${v}c$vd e"}]}`: {`{"a":[true,{"b":"a1c e"}]}`, []string{"v", "vd"}},
		`{"a": [true, {"b": {"c": "a${v}c$v"}}]}`: {`{"a":[true,{"b":{"c":"a1c1"}}]}`, []string{"v"}},
		`"a${v}${v1}c"`: {`"a1c"`, []string{"v", "v1"}},
		`{"a": "$v"}`:   {`{"a":"1"}`, []string{"v"}},
		`{"a": "$$v"}`:  {`{"a":"$v"}`, []string{}},
	}
	for data, want := range tests {
		t.Run(data, func(t *testing.T) {
			got, seenVars, err := expand([]byte(data), mapping)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != want.interpolatedData {
				t.Errorf("got data != want data\n got %s\nwant %s", got, want.interpolatedData)
			}
			if !reflect.DeepEqual(seenVars, want.seenVars) {
				t.Errorf("got seenVars %q, want %q", seenVars, want.seenVars)
			}
		})
	}
}
