package backend

import (
	"reflect"
	"strings"
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

// TestSearch tests the search endpoint with mocks for store.Defs (used to get
// the definition IDs to return in search results) and Graph.Defs (used to fill
// in the definition metadata for each result).
func TestSearch(t *testing.T) {
	const (
		repoURI = "r1"
		repoID  = 1
	)

	ctx, mock := testContext()

	s := search{}

	// Mock data
	wantDefsKeys := []graph.DefKey{
		{Repo: repoURI, CommitID: "c1", UnitType: "t1", Unit: "u1", Path: "p1"},
		{Repo: repoURI, CommitID: "c1", UnitType: "t1", Unit: "u1", Path: "p2"},
		{Repo: repoURI, CommitID: "c2", UnitType: "t2", Unit: "u2", Path: "p3"},
	}
	wantDefResults := make([]*sourcegraph.DefSearchResult, len(wantDefsKeys))
	for i, dk := range wantDefsKeys {
		d := &sourcegraph.Def{Def: graph.Def{DefKey: dk}}
		wantDefResults[i] = &sourcegraph.DefSearchResult{Def: *d, Score: 0, RefCount: 0}
	}
	wantResults := &sourcegraph.SearchResultsList{
		DefResults:         wantDefResults,
		SearchQueryOptions: []*sourcegraph.SearchOptions{{ListOptions: sourcegraph.ListOptions{}}},
	}

	query := "this is the test query"
	expTokQuery := strings.Fields(query)

	calledDefsSearch := false
	mock.stores.Defs.Search_ = func(ctx context.Context, op store.DefSearchOp) (*sourcegraph.SearchResultsList, error) {
		calledDefsSearch = true
		if !reflect.DeepEqual(expTokQuery, op.TokQuery) {
			t.Fatalf("expected op.TokQuery %+v, got %+v", expTokQuery, op.TokQuery)
		}
		return &sourcegraph.SearchResultsList{DefResults: wantDefResults}, nil
	}

	// Test that search endpoint works.
	results, err := s.Search(ctx, &sourcegraph.SearchOp{
		Query: query,
		Opt:   &sourcegraph.SearchOptions{},
	})
	if err != nil {
		t.Fatalf("unexpected error from search.Search: %s", err)
	}

	// Test that backend stores were called
	if !reflect.DeepEqual(wantResults, results) {
		t.Errorf("wanted\n%+v, got\n%+v", wantResults, results)
	}

	if !calledDefsSearch {
		t.Errorf("!calledDefsSearch")
	}
}
