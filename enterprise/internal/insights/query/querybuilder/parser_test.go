package querybuilder

import (
	"github.com/google/go-cmp/cmp"
	"sort"
	"testing"
)

func TestParseAndValidateQuery(t *testing.T) {
	testCases := []struct {
		name  string
		query string
		fail  bool
	}{
		{
			"invalid parameter type",
			"select:repo test fork:only.",
			true,
		},
		{
			"valid query",
			"select:file test",
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hasFailed := false
			_, err := ParseAndValidateQuery(tc.query, "literal")
			if err != nil {
				hasFailed = true
			}
			if tc.fail != hasFailed {
				t.Errorf("expected %v result, got %v", tc.fail, hasFailed)
			}
		})
	}
}

func TestParametersFromQueryPlan(t *testing.T) {
	testCases := []struct {
		name       string
		query      string
		parameters []string
	}{
		{
			"returns single parameter",
			"select:repo",
			[]string{`"select:repo"`},
		},
		{
			"returns multiple parameters",
			"select:file file:insights test",
			[]string{`"file:insights"`, `"select:file"`},
		},
		{
			"returns no parameter",
			"I am search",
			[]string{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := ParseAndValidateQuery(tc.query, "literal")
			if err != nil {
				t.Errorf("expected valid query, got error: %v", err)
			}
			parameterStrings := []string{}
			for _, parameter := range ParametersFromQueryPlan(plan) {
				parameterStrings = append(parameterStrings, parameter.String())
			}
			sort.Strings(parameterStrings)
			if diff := cmp.Diff(parameterStrings, tc.parameters); diff != "" {
				t.Errorf("expected %v, got %v", tc.parameters, parameterStrings)
			}
		})
	}
}
