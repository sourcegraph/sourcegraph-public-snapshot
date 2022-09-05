package querybuilder

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

func TestParseQuery(t *testing.T) {
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
		{
			"valid literal query",
			"select:file i++",
			false,
		},
		{
			"invalid regexp query submitted as literal",
			"patterntype:regexp i++",
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hasFailed := false
			_, err := ParseQuery(tc.query, "literal")
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
			plan, err := ParseQuery(tc.query, "literal")
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

func TestDetectSearchType(t *testing.T) {
	testCases := []struct {
		name          string
		query         string
		submittedType string
		searchType    query.SearchType
	}{
		{
			"submitted and query match types",
			"select:repo test fork:only",
			"literal",
			query.SearchTypeLiteral,
		},
		{
			"submit literal with patterntype",
			"test patterntype:regexp",
			"literal",
			query.SearchTypeRegex,
		},
		{
			"submit literal with patterntype",
			"test patterntype:regexp",
			"lucky",
			query.SearchTypeRegex,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			searchType, err := DetectSearchType(tc.query, tc.submittedType)
			if err != nil {
				t.Errorf("expected %d, errored: %s", tc.searchType, err.Error())
			}
			if tc.searchType != searchType {
				t.Errorf("expected %d result, got %d", tc.searchType, searchType)
			}
		})
	}
}

func TestRepoOnlyQuery(t *testing.T) {
	testCases := []struct {
		name        string
		query       string
		patternType string
		repoOnly    bool
	}{
		{
			"select:repo parameter",
			"select:repo",
			"literal",
			true,
		},
		{
			"repo parameter",
			"repo:sourcegraph",
			"literal",
			true,
		},
		{
			"multiple repo filters",
			"repo:sourcegraph repo:k8s select:repo",
			"literal",
			true,
		},
		{
			"regexp repo filter",
			"repo:deploy-\\w+",
			"regexp",
			true,
		},
		{
			"repo or query",
			"repo:has.path(README) or repo:sourcegraph",
			"literal",
			true,
		},
		{
			"repo and query",
			"repo:has.path(README) and repo:sourcegraph",
			"literal",
			true,
		},
		{
			"repo not query",
			"not repo:sg",
			"literal",
			true,
		},
		{
			"negated repo",
			"-repo:sg",
			"literal",
			true,
		},
		{
			"query with file parameter",
			"repo:sg file:insights",
			"literal",
			false,
		},
		{
			"query with search pattern",
			"repo:sg test",
			"literal",
			false,
		},
		{
			"repo and search pattern",
			"repo:sg and test",
			"literal",
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := ParseQuery(tc.query, tc.patternType)
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			got := RepoOnlyQuery(plan)
			if got != tc.repoOnly {
				t.Errorf("expected %v, got %v", tc.repoOnly, got)
			}
		})
	}
}
