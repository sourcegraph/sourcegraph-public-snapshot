package protocol_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/searcher/protocol"
)

func TestFromJobNode(t *testing.T) {
	cases := []struct {
		name       string
		query      string
		searchType query.SearchType
		want       protocol.QueryNode
	}{{
		name:       "single pattern",
		query:      "foobar",
		searchType: query.SearchTypeStandard,
		want:       &protocol.PatternNode{Value: "foobar"},
	}, {
		name:       "AND query",
		query:      "foo AND bar",
		searchType: query.SearchTypeStandard,
		want: &protocol.AndNode{
			Children: []protocol.QueryNode{
				&protocol.PatternNode{Value: "foo"},
				&protocol.PatternNode{Value: "bar"},
			},
		},
	}, {
		name:       "complex query with negation",
		query:      "(foo AND NOT bar) OR NOT baz OR buzz",
		searchType: query.SearchTypeStandard,
		want: &protocol.OrNode{
			Children: []protocol.QueryNode{
				&protocol.AndNode{
					Children: []protocol.QueryNode{
						&protocol.PatternNode{Value: "foo"},
						&protocol.PatternNode{Value: "bar", IsNegated: true},
					},
				},
				&protocol.PatternNode{Value: "baz", IsNegated: true},
				&protocol.PatternNode{Value: "buzz"},
			},
		},
	}, {
		name:       "complex query with regex",
		query:      "(foo AND NOT /bar?/) OR NOT baz OR /.*buzz/",
		searchType: query.SearchTypeStandard,
		want: &protocol.OrNode{
			Children: []protocol.QueryNode{
				&protocol.AndNode{
					Children: []protocol.QueryNode{
						&protocol.PatternNode{Value: "foo"},
						&protocol.PatternNode{Value: "bar?", IsNegated: true, IsRegExp: true},
					},
				},
				&protocol.PatternNode{Value: "baz", IsNegated: true},
				&protocol.PatternNode{Value: ".*buzz", IsRegExp: true},
			},
		},
	}, {
		name:       "regex pattern type",
		query:      "(foo AND NOT bar?) OR NOT baz OR .*buzz",
		searchType: query.SearchTypeRegex,
		want: &protocol.OrNode{
			Children: []protocol.QueryNode{
				&protocol.AndNode{
					Children: []protocol.QueryNode{
						&protocol.PatternNode{Value: "foo", IsRegExp: true},
						&protocol.PatternNode{Value: "bar?", IsNegated: true, IsRegExp: true},
					},
				},
				&protocol.PatternNode{Value: "baz", IsNegated: true, IsRegExp: true},
				&protocol.PatternNode{Value: ".*buzz", IsRegExp: true},
			},
		},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nodes, err := query.Parse(tc.query, tc.searchType)
			if err != nil {
				t.Fatal(err)
			}

			if len(nodes) != 1 {
				t.Fatal("invalid test case: expected query to parse to 1 node")
			}

			got := protocol.FromJobNode(nodes[0])
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatalf("expected queries to be equal. diff: %s", diff)
			}
		})
	}
}

func TestString(t *testing.T) {
	q := &protocol.OrNode{
		Children: []protocol.QueryNode{
			&protocol.AndNode{
				Children: []protocol.QueryNode{
					&protocol.PatternNode{Value: "foo"},
					&protocol.PatternNode{Value: "bar", IsNegated: true},
				},
			},
			&protocol.PatternNode{Value: "baz", IsNegated: true},
			&protocol.PatternNode{Value: "buzz", IsRegExp: true},
		},
	}

	want := `(("foo" AND NOT "bar") OR NOT "baz" OR /buzz/)`
	if diff := cmp.Diff(want, q.String()); diff != "" {
		t.Fatalf("expected string output to be equal. diff: %s", diff)
	}
}
