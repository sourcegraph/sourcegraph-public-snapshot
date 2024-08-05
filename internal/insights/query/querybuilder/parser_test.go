package querybuilder

import (
	"fmt"
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
			"submit structural with structural patterntype",
			"[a] patterntype:structural",
			"structural",
			query.SearchTypeStructural,
		},
		{
			"submit regexp with structural patterntype",
			"[a] patterntype:regexp",
			"structural",
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

func TestContainsField(t *testing.T) {
	testCases := []struct {
		name  string
		query string
		field string
		found bool
	}{
		{
			"field not present",
			"select:repo",
			query.FieldRepo,
			false,
		},
		{
			"field present",
			"select:file repo:test",
			query.FieldRepo,
			true,
		},
		{
			"field multiple times",
			"(file:test repo:test) OR (file:nottest repo:nottest)",
			query.FieldRepo,
			true,
		},
		{
			"finds alias",
			"r:test thing",
			query.FieldRepo,
			true,
		},
		{
			"does not false positive",
			`file:test content:"repo:"`,
			query.FieldRepo,
			false,
		},
		{
			"is not case sensitive",
			`rEpO:test my search`,
			query.FieldRepo,
			true,
		},
		{
			"field in first plan of query",
			"(file:test repo:test) OR (some other search)",
			query.FieldRepo,
			true,
		},
		{
			"field in 2nd plan of query",
			"(some other search) OR (file:test repo:test) ",
			query.FieldRepo,
			true,
		},
		{
			"doesn't count empty field",
			"mysearch repo:",
			query.FieldRepo,
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			found, err := ContainsField(tc.query, tc.field)
			if err != nil {
				t.Errorf("expected valid query, got error: %v", err)
			}
			if diff := cmp.Diff(found, tc.found); diff != "" {
				t.Errorf("expected %v, got %v", tc.found, found)
			}
		})
	}
}

func TestIsValidScopeQuery(t *testing.T) {
	testCases := []struct {
		name   string
		query  string
		valid  bool
		reason string
	}{
		{
			name:   "invalid single query with pattern",
			query:  "repo:sourcegraph pattern",
			valid:  false,
			reason: fmt.Sprintf(containsPattern, "pattern"),
		},
		{
			name:   "invalid multiple query with pattern",
			query:  "repo:sourcegraph or repo:about pattern",
			valid:  false,
			reason: fmt.Sprintf(containsPattern, "pattern"),
		},
		{
			name:   "invalid query with disallowed filter",
			query:  "file:sourcegraph repo:handbook",
			valid:  false,
			reason: fmt.Sprintf(containsDisallowedFilter, "file"),
		},
		{
			name:   "invalid query with Uppercase filter",
			query:  "REpo:sourcegraph or lang:go",
			valid:  false,
			reason: fmt.Sprintf(containsDisallowedFilter, "lang"),
		},
		{
			name:  "valid multiple query",
			query: "repo:sourcegraph or repo:about and repo:handbook",
			valid: true,
		},
		{
			name:  "valid query with shorthand repo filter",
			query: "r:sourcegraph",
			valid: true,
		},
		{
			name:  "valid query with repo predicate filter",
			query: "repo:has.file(path:README)",
			valid: true,
		},
		{
			name:   "invalid query with rev filter",
			query:  "repo:sourcegraph rev:mybranch",
			reason: containsDisallowedRevision,
			valid:  false,
		},
		{
			name:   "invalid query with specified on repo filter",
			query:  `repo:^github\.com/sourcegraph/sourcegraph$@v4.0.0`,
			reason: containsDisallowedRevision,
			valid:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := ParseQuery(tc.query, "literal")
			if err != nil {
				t.Fatal(err)
			}
			reason, valid := IsValidScopeQuery(plan)
			if valid != tc.valid {
				t.Errorf("expected validity %v, got %v", tc.valid, valid)
			}
			if reason != tc.reason {
				t.Errorf("expected reason %v, got %v", tc.reason, reason)
			}
		})
	}
}
