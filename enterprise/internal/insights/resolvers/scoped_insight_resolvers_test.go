package resolvers

import (
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
)

func Test_isValidScopeQuery(t *testing.T) {
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
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := querybuilder.ParseQuery(tc.query, "literal")
			if err != nil {
				t.Fatal(err)
			}
			reason, valid := isValidScopeQuery(plan)
			if valid != tc.valid {
				t.Errorf("expected validity %v, got %v", tc.valid, valid)
			}
			if reason != tc.reason {
				t.Errorf("expected reason %v, got %v", tc.reason, reason)
			}
		})
	}
}
