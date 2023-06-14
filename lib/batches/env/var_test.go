package env

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestVariable_MarshalJSON(t *testing.T) {
	for name, tc := range map[string]struct {
		in   variable
		want string
	}{
		"no value": {
			in:   variable{name: "foo"},
			want: `"foo"`,
		},
		"with value": {
			in:   variable{name: "foo", value: pointers.Ptr("bar")},
			want: `{"foo":"bar"}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			have, err := json.Marshal(tc.in)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if string(have) != tc.want {
				t.Errorf("unexpected value: have=%q want=%q", have, tc.want)
			}
		})
	}
}

func TestVariable_UnmarshalJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			in   string
			want variable
		}{
			"no value": {
				in:   `"foo"`,
				want: variable{name: "foo"},
			},
			"with value": {
				in:   `{"foo":"bar"}`,
				want: variable{name: "foo", value: pointers.Ptr("bar")},
			},
		} {
			t.Run(name, func(t *testing.T) {
				var have variable
				if err := json.Unmarshal([]byte(tc.in), &have); err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if diff := cmp.Diff(have, tc.want); diff != "" {
					t.Errorf("unexpected value:\n%s", diff)
				}
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		t.Run("invalid types", func(t *testing.T) {
			for name, in := range map[string]string{
				"invalid outer type": `false`,
				"invalid inner type": `{"foo":false}`,
			} {
				t.Run(name, func(t *testing.T) {
					var have variable
					if err := json.Unmarshal([]byte(in), &have); err == nil {
						t.Error("unexpected nil error")
					} else if err != errInvalidVariableType {
						t.Errorf("unexpected error: have=%v want=%v", err, errInvalidVariableType)
					}
				})
			}
		})

		t.Run("invalid objects", func(t *testing.T) {
			for name, tc := range map[string]struct {
				in   string
				want int
			}{
				"no properties": {
					in:   `{}`,
					want: 0,
				},
				"too many properties": {
					in:   `{"a":"b","c":"d"}`,
					want: 2,
				},
			} {
				t.Run(name, func(t *testing.T) {
					var have variable
					if err := json.Unmarshal([]byte(tc.in), &have); err == nil {
						t.Error("unexpected nil error")
					} else if e, ok := err.(errInvalidVariableObject); !ok {
						t.Errorf("unexpected error of type %T: %v", err, err)
					} else if e.n != tc.want {
						t.Errorf("unexpected number of properties in the error: have=%d want=%d", e.n, tc.want)
					} else if e.Error() == "" {
						t.Error("unexpected empty error string")
					}
				})
			}
		})
	})
}

func TestVariable_UnmarshalYAML(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			in   string
			want variable
		}{
			"no value": {
				in:   `foo`,
				want: variable{name: "foo"},
			},
			"with value": {
				in:   `foo: bar`,
				want: variable{name: "foo", value: pointers.Ptr("bar")},
			},
		} {
			t.Run(name, func(t *testing.T) {
				var have variable
				if err := yaml.Unmarshal([]byte(tc.in), &have); err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if diff := cmp.Diff(have, tc.want); diff != "" {
					t.Errorf("unexpected value:\n%s", diff)
				}
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		t.Run("invalid types", func(t *testing.T) {
			for name, in := range map[string]string{
				"invalid outer type": `[]`,
				"invalid inner type": `foo: []`,
			} {
				t.Run(name, func(t *testing.T) {
					var have variable
					if err := yaml.Unmarshal([]byte(in), &have); err == nil {
						t.Error("unexpected nil error")
					} else if err != errInvalidVariableType {
						t.Errorf("unexpected error: have=%v want=%v", err, errInvalidVariableType)
					}
				})
			}
		})

		t.Run("invalid objects", func(t *testing.T) {
			for name, tc := range map[string]struct {
				in   string
				want int
			}{
				"no properties": {
					in:   `{}`,
					want: 0,
				},
				"too many properties": {
					in:   "a: b\nc: d",
					want: 2,
				},
			} {
				t.Run(name, func(t *testing.T) {
					var have variable
					if err := yaml.Unmarshal([]byte(tc.in), &have); err == nil {
						t.Error("unexpected nil error")
					} else if e, ok := err.(errInvalidVariableObject); !ok {
						t.Errorf("unexpected error of type %T: %v", err, err)
					} else if e.n != tc.want {
						t.Errorf("unexpected number of properties in the error: have=%d want=%d", e.n, tc.want)
					} else if e.Error() == "" {
						t.Error("unexpected empty error string")
					}
				})
			}
		})
	})
}
