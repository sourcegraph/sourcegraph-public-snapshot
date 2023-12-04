package env

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestEnvironment_MarshalJSON(t *testing.T) {
	for name, tc := range map[string]struct {
		in   Environment
		want string
	}{
		"no variables": {
			in:   Environment{},
			want: `{}`,
		},
		"only static variables": {
			in: Environment{vars: []variable{
				{name: "foo", value: pointers.Ptr("bar")},
				{name: "quux", value: pointers.Ptr("baz")},
			}},
			want: `{"foo":"bar","quux":"baz"}`,
		},
		"with variables": {
			in: Environment{vars: []variable{
				{name: "foo", value: pointers.Ptr("bar")},
				{name: "quux", value: nil},
			}},
			want: `[{"foo":"bar"},"quux"]`,
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

func TestEnvironment_UnmarshalJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			in   string
			want Environment
		}{
			"empty array": {
				in:   `[]`,
				want: Environment{},
			},
			"set array": {
				in: `[{"foo":"bar"},"quux"]`,
				want: Environment{vars: []variable{
					{name: "foo", value: pointers.Ptr("bar")},
					{name: "quux"},
				}},
			},
			"empty object": {
				in:   `{}`,
				want: Environment{},
			},
			"set object": {
				in: `{"foo":"bar","quux":"baz"}`,
				want: Environment{vars: []variable{
					{name: "foo", value: pointers.Ptr("bar")},
					{name: "quux", value: pointers.Ptr("baz")},
				}},
			},
		} {
			t.Run(name, func(t *testing.T) {
				var have Environment
				if err := json.Unmarshal([]byte(tc.in), &have); err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if diff := cmp.Diff(have, tc.want); diff != "" {
					t.Errorf("unexpected environment:\n%s", diff)
				}
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		for name, in := range map[string]string{
			"invalid outer type":             `false`,
			"invalid object inner type":      `{"foo":false}`,
			"invalid array inner type":       `[false]`,
			"invalid array inner inner type": `[{"foo":false}]`,
		} {
			t.Run(name, func(t *testing.T) {
				var have Environment
				if err := json.Unmarshal([]byte(in), &have); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})
}

func TestEnvironment_UnmarshalYAML(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			in   string
			want Environment
		}{
			"empty array": {
				in:   `[]`,
				want: Environment{},
			},
			"set array": {
				in: "- foo: bar\n- quux",
				want: Environment{vars: []variable{
					{name: "foo", value: pointers.Ptr("bar")},
					{name: "quux"},
				}},
			},
			"empty object": {
				in:   `{}`,
				want: Environment{},
			},
			"set object": {
				in: "foo: bar\nquux: baz",
				want: Environment{vars: []variable{
					{name: "foo", value: pointers.Ptr("bar")},
					{name: "quux", value: pointers.Ptr("baz")},
				}},
			},
		} {
			t.Run(name, func(t *testing.T) {
				var have Environment
				if err := yaml.Unmarshal([]byte(tc.in), &have); err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if diff := cmp.Diff(have, tc.want); diff != "" {
					t.Errorf("unexpected environment:\n%s", diff)
				}
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		for name, in := range map[string]string{
			"invalid outer type":             `foo`,
			"invalid object inner type":      `foo: []`,
			"invalid array inner type":       `[[]]`,
			"invalid array inner inner type": `[{"foo":[]]}]`,
		} {
			t.Run(name, func(t *testing.T) {
				var have Environment
				if err := yaml.Unmarshal([]byte(in), &have); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})
}

func TestEnvironment_IsStatic(t *testing.T) {
	for name, tc := range map[string]struct {
		env  Environment
		want bool
	}{
		"empty": {
			env:  Environment{},
			want: true,
		},
		"static": {
			env: Environment{vars: []variable{
				{name: "foo", value: pointers.Ptr("bar")},
				{name: "quux", value: pointers.Ptr("baz")},
			}},
			want: true,
		},
		"not static": {
			env: Environment{vars: []variable{
				{name: "foo", value: pointers.Ptr("bar")},
				{name: "quux", value: nil},
			}},
			want: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if have := tc.env.IsStatic(); have != tc.want {
				t.Errorf("unexpected static value: have=%v want=%v", have, tc.want)
			}
		})
	}
}

func TestEnvironment_Resolve(t *testing.T) {
	env := Environment{vars: []variable{
		{name: "nil"},
		{name: "foo", value: pointers.Ptr("bar")},
	}}

	t.Run("invalid outer", func(t *testing.T) {
		if _, err := env.Resolve([]string{"foo"}); err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("valid", func(t *testing.T) {
		for name, tc := range map[string]struct {
			outer []string
			want  map[string]string
		}{
			"nil outer": {
				outer: nil,
				want:  map[string]string{"nil": "", "foo": "bar"},
			},
			"empty outer": {
				outer: []string{},
				want:  map[string]string{"nil": "", "foo": "bar"},
			},
			"outer doesn't fill in value": {
				outer: []string{"quux=baz"},
				want:  map[string]string{"nil": "", "foo": "bar"},
			},
			"outer does fill in empty value": {
				outer: []string{"nil="},
				want:  map[string]string{"nil": "", "foo": "bar"},
			},
			"outer does fill in value": {
				outer: []string{"nil=baz"},
				want:  map[string]string{"nil": "baz", "foo": "bar"},
			},
			"outer does fill in value with equal sign": {
				outer: []string{"nil=baz=fuzz"},
				want:  map[string]string{"nil": "baz=fuzz", "foo": "bar"},
			},
			"outer also contains value not to be filled in": {
				outer: []string{"nil=baz", "foo=not bar"},
				want:  map[string]string{"nil": "baz", "foo": "bar"},
			},
			"outer also contains empty value not to be filled in": {
				outer: []string{"nil=baz", "foo="},
				want:  map[string]string{"nil": "baz", "foo": "bar"},
			},
		} {
			t.Run(name, func(t *testing.T) {
				if have, err := env.Resolve(tc.outer); err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if diff := cmp.Diff(have, tc.want); diff != "" {
					t.Errorf("unexpected resolved environment:\n%s", diff)
				}
			})
		}
	})
}

func TestEnvironment_OuterVars(t *testing.T) {
	for name, tc := range map[string]struct {
		in   Environment
		want []string
	}{
		"no variables": {
			in:   Environment{},
			want: []string{},
		},
		"static variables": {
			in: Environment{vars: []variable{
				{name: "foo", value: pointers.Ptr("bar")},
				{name: "quux", value: pointers.Ptr("baz")},
			}},
			want: []string{},
		},
		"dynamic variables and static mixed": {
			in: Environment{vars: []variable{
				{name: "foo", value: pointers.Ptr("bar")},
				{name: "quux", value: nil},
			}},
			want: []string{"quux"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			have := tc.in.OuterVars()

			if diff := cmp.Diff(have, tc.want); diff != "" {
				t.Errorf("unexpected value: have=%q want=%q", have, tc.want)
			}
		})
	}
}
