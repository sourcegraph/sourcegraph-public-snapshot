package api

import "testing"

func TestJSONValueWithFields(t *testing.T) {
	tests := map[string]struct {
		jsoncStr string
		fields   []string
		want     string
	}{
		"invalid json top-level": {
			jsoncStr: `{`,
			fields:   []string{"x"},
			want:     `{}`,
		},
		"invalid json in field": {
			jsoncStr: `{"a": {"b": }, "c": 3}`,
			fields:   []string{"a", "c"},
			want:     `{}`,
		},
		"subset of fields": {
			jsoncStr: `{"a": 1, "b": {"c": 3,}, "d": true,}`,
			fields:   []string{"d", "b", "x"},
			want:     `{"b":{"c":3},"d":true}`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := jsonValueWithFields(test.jsoncStr, test.fields)
			got, err := v.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != test.want {
				t.Errorf("got %s, want %s", got, test.want)
			}
		})
	}
}
