package overridable

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

func TestBoolIs(t *testing.T) {
	for name, tc := range map[string]struct {
		in   Bool
		name string
		want bool
	}{
		"wildcard false": {
			in: Bool{
				rules: rules{{pattern: allPattern, value: false}},
			},
			name: "foo",
			want: false,
		},
		"wildcard true": {
			in: Bool{
				rules: rules{{pattern: allPattern, value: true}},
			},
			name: "foo",
			want: true,
		},
		"list exhausted": {
			in: Bool{
				rules: rules{{pattern: "bar*", value: true}},
			},
			name: "foo",
			want: false,
		},
		"single match": {
			in: Bool{
				rules: rules{{pattern: "bar*", value: true}},
			},
			name: "bar",
			want: true,
		},
		"multiple matches": {
			in: Bool{
				rules: rules{
					{pattern: allPattern, value: true},
					{pattern: "bar*", value: false},
				},
			},
			name: "bar",
			want: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := initBool(&tc.in); err != nil {
				t.Fatal(err)
			}

			if have := tc.in.Value(tc.name); have != tc.want {
				t.Errorf("unexpected value: have=%v want=%v", have, tc.want)
			}
		})
	}
}

func TestBoolMarshalJSON(t *testing.T) {
	bs := Bool{
		rules{
			{pattern: allPattern, value: true},
			{pattern: "bar*", value: false},
		},
	}
	data, err := json.Marshal(&bs)
	if err != nil {
		t.Errorf("unexpected non-nil error: %v", err)
	}
	if have, want := string(data), `[{"*":true},{"bar*":false}]`; have != want {
		t.Errorf("unexpected JSON: have=%q want=%q", have, want)
	}
}

func TestBoolUnmarshalJSON(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for name, tc := range map[string]struct {
			in   string
			want Bool
		}{
			"single false": {
				in: `false`,
				want: Bool{
					rules: rules{
						{pattern: allPattern, value: false},
					},
				},
			},
			"single true": {
				in: `true`,
				want: Bool{
					rules: rules{
						{pattern: allPattern, value: true},
					},
				},
			},
			"empty list": {
				in: `[]`,
				want: Bool{
					rules: rules{},
				},
			},
			"multiple rule list": {
				in: `[{"*":true},{"github.com/sourcegraph/*":false}]`,
				want: Bool{
					rules: rules{
						{pattern: allPattern, value: true},
						{pattern: "github.com/sourcegraph/*", value: false},
					},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				var have Bool
				if err := json.Unmarshal([]byte(tc.in), &have); err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}
				if diff := cmp.Diff(&have, &tc.want); diff != "" {
					t.Errorf("unexpected Bool: %s", diff)
				}
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for name, in := range map[string]string{
			"string":          `"foo"`,
			"empty object":    `[{}]`,
			"too many fields": `[{"foo": true,"bar":false}]`,
			"invalid glob":    `[{"[":false}]`,
		} {
			t.Run(name, func(t *testing.T) {
				var have Bool
				if err := json.Unmarshal([]byte(in), &have); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})
}

func TestBoolYAML(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for name, tc := range map[string]struct {
			in   string
			want Bool
		}{
			"single false": {
				in: `false`,
				want: Bool{
					rules: rules{
						{pattern: allPattern, value: false},
					},
				},
			},
			"single true": {
				in: `true`,
				want: Bool{
					rules: rules{
						{pattern: allPattern, value: true},
					},
				},
			},
			"empty list": {
				in: `[]`,
				want: Bool{
					rules: rules{},
				},
			},
			"multiple rule list": {
				in: "- \"*\": true\n- github.com/sourcegraph/*: false",
				want: Bool{
					rules: rules{
						{pattern: allPattern, value: true},
						{pattern: "github.com/sourcegraph/*", value: false},
					},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				var have Bool
				if err := yaml.Unmarshal([]byte(tc.in), &have); err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}
				if diff := cmp.Diff(&have, &tc.want); diff != "" {
					t.Errorf("unexpected Bool: %s", diff)
				}
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for name, in := range map[string]string{
			"string":          `foo`,
			"empty object":    `- {}`,
			"too many fields": `- {"foo": true, "bar": false}`,
			"invalid glob":    `- "[": false`,
		} {
			t.Run(name, func(t *testing.T) {
				var have Bool
				if err := yaml.Unmarshal([]byte(in), &have); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})
}

// initBool ensures all rules are compiled.
func initBool(b *Bool) (err error) {
	for i, rule := range b.rules {
		if rule.compiled == nil {
			b.rules[i], err = newRule(rule.pattern, rule.value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
