package backend

import (
	"reflect"
	"strings"
	"sync"
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
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
	wantDefs := make([]*sourcegraph.Def, len(wantDefsKeys))
	var (
		calledDefGetMu sync.Mutex
		calledDefGet   = make(map[graph.DefKey]bool)
	)
	for i, dk := range wantDefsKeys {
		wantDefs[i] = &sourcegraph.Def{Def: graph.Def{DefKey: dk}}
		calledDefGet[dk] = false
	}
	wantDefResults := make([]*sourcegraph.DefSearchResult, len(wantDefs))
	for i, d := range wantDefs {
		wantDefResults[i] = &sourcegraph.DefSearchResult{Def: *d, Score: 0, RefCount: 0}
	}
	wantResults := &sourcegraph.SearchResultsList{
		DefResults:         wantDefResults,
		SearchQueryOptions: []*sourcegraph.SearchOptions{{ListOptions: sourcegraph.ListOptions{}}},
	}

	query := "this is the test query"
	expTokQuery := strings.Fields(query)

	// Declare mocks
	mock.servers.Defs.Get_ = func(ctx context.Context, op *sourcegraph.DefsGetOp) (*sourcegraph.Def, error) {
		calledDefGetMu.Lock()
		defer calledDefGetMu.Unlock()

		if op.Def.Repo != repoID {
			t.Fatalf("expected request for def from repo %d, got %d", repoID, op.Def.Repo)
		}
		// Mock fetch of defs (should be called to hydrate each def result)
		for _, d := range wantDefs {
			if d.CommitID == op.Def.CommitID && d.Unit == op.Def.Unit && d.UnitType == op.Def.UnitType && d.Path == op.Def.Path {
				calledDefGet[graph.DefKey{Repo: repoURI, CommitID: d.CommitID, Unit: d.Unit, UnitType: d.UnitType, Path: d.Path}] = true
				return d, nil
			}
		}
		t.Fatalf("attempted to request unmocked def: %+v", op)
		return nil, nil
	}
	mock.stores.Repos.GetByURI = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		return &sourcegraph.Repo{URI: repo, ID: repoID}, nil
	}

	calledDefsSearch := false
	mock.stores.Defs.Search = func(ctx context.Context, op localstore.DefSearchOp) (*sourcegraph.SearchResultsList, error) {
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

	for dk, calledGet := range calledDefGet {
		if !calledGet {
			t.Errorf("failed to call Graph.Defs.Get on def %+v", dk)
		}
	}
}
