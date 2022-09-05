package overridable

import (
	"encoding/json"
	"testing"
)

func TestRuleInvalid(t *testing.T) {
	if _, err := newRule("[", true); err == nil {
		t.Error("unexpected nil error")
	}
}

func TestRulesMarshalJSON(t *testing.T) {
	for name, tc := range map[string]struct {
		in   rules
		want string
	}{
		"no rules": {
			in:   rules{},
			want: `[]`,
		},
		"one wildcard rule": {
			in:   rules{{pattern: allPattern, value: true}},
			want: `true`,
		},
		"one non-wildcard rule": {
			in:   rules{{pattern: "bar*", value: true}},
			want: `[{"bar*":true}]`,
		},
		"multiple rules": {
			in: rules{
				{pattern: allPattern, value: true},
				{pattern: "bar*", value: false},
				{pattern: "foo*", value: "draft"},
			},
			want: `[{"*":true},{"bar*":false},{"foo*":"draft"}]`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(&tc.in)
			if err != nil {
				t.Errorf("unexpected non-nil error: %v", err)
			}
			if string(data) != tc.want {
				t.Errorf("unexpected JSON: have=%q want=%q", string(data), tc.want)
			}
		})
	}
}

func TestMatchWithSuffix(t *testing.T) {
	type ruleInputs struct {
		pattern string
		value   any
	}
	compileInputs := func(t *testing.T, inputs []ruleInputs) (rs rules) {
		for _, input := range inputs {
			r, err := newRule(input.pattern, input.value)
			if err != nil {
				t.Fatalf("failed to compile rule. pattern=%q, value=%+v", input.pattern, input.value)
			}
			rs = append(rs, r)
		}
		return
	}

	for name, tc := range map[string]struct {
		rules []ruleInputs
		args  []string
		want  any
	}{
		"no rules": {
			rules: []ruleInputs{},
			args:  []string{"name", "suffix"},
			want:  nil,
		},
		"no match": {
			rules: []ruleInputs{
				{pattern: "repo*@branch-name-1", value: "rule-1"},
				{pattern: "repo*@branch-name-2", value: "rule-2"},
			},
			args: []string{"repo-1000", "other-branch-name"},
			want: nil,
		},
		"single match": {
			rules: []ruleInputs{
				{pattern: "repo*@other-branch-name", value: "rule-1"},
				{pattern: "repo*@branch-name", value: "rule-2"},
			},
			args: []string{"repo-1000", "other-branch-name"},
			want: "rule-1",
		},
		"multiple matches": {
			rules: []ruleInputs{
				{pattern: "repo*@branch-name", value: "rule-1"},
				{pattern: "repo*@branch-name", value: "rule-2"},
			},
			args: []string{"repo-1000", "branch-name"},
			want: "rule-2",
		},
	} {
		t.Run(name, func(t *testing.T) {
			rs := compileInputs(t, tc.rules)

			have := rs.MatchWithSuffix(tc.args[0], tc.args[1])
			if have != tc.want {
				t.Errorf("unexpected match. want=%+v, have=%+v", tc.want, have)
			}
		})
	}
}
