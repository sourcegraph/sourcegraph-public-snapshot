package symbols

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSearchWithFiltering(t *testing.T) {
	ctx := context.Background()
	fixture := search.SymbolsResponse{
		Symbols: result.Symbols{
			result.Symbol{
				Name: "foo1",
				Path: "file1",
			},
			result.Symbol{
				Name: "foo2",
				Path: "file2",
			},
		}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(fixture)
	}))
	t.Cleanup(func() {
		srv.Close()
	})
	DefaultClient.URL = srv.URL

	results, err := DefaultClient.Search(ctx, search.SymbolsParameters{
		Repo:     "foo",
		CommitID: "HEAD",
		Query:    "abc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if results == nil {
		t.Fatal("nil result")
	}
	wantCount := 2
	if len(results) != wantCount {
		t.Fatalf("Want %d results, got %d", wantCount, len(results))
	}

	// With filtering
	ctx = actor.WithActor(ctx, &actor.Actor{
		UID: 1,
	})
	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
		if content.Path == "file1" {
			return authz.Read, nil
		}
		return authz.None, nil
	})
	authz.DefaultSubRepoPermsChecker = checker

	results, err = DefaultClient.Search(ctx, search.SymbolsParameters{
		Repo:     "foo",
		CommitID: "HEAD",
		Query:    "abc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if results == nil {
		t.Fatal("nil result")
	}
	wantCount = 1
	if len(results) != wantCount {
		t.Fatalf("Want %d results, got %d", wantCount, len(results))
	}
}

func TestDefinitionWithFiltering(t *testing.T) {
	// This test conflicts with the previous use of httptest.NewServer, but passes in isolation.
	t.Skip()

	path1 := types.RepoCommitPathPoint{
		RepoCommitPath: types.RepoCommitPath{
			Repo:   "somerepo",
			Commit: "somecommit",
			Path:   "path1",
		},
		Point: types.Point{Row: 0, Column: 0},
	}

	path2 := types.RepoCommitPathPoint{
		RepoCommitPath: types.RepoCommitPath{
			Repo:   "somerepo",
			Commit: "somecommit",
			Path:   "path2",
		},
		Point: types.Point{Row: 0, Column: 0},
	}

	// Start an HTTP server that responds with path1.
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(path1)
	}))
	t.Cleanup(func() {
		srv.Close()
	})
	DefaultClient.URL = srv.URL

	// Request path1.
	results, err := DefaultClient.SymbolInfo(ctx, path2)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure we get results.
	if results == nil {
		t.Fatal("nil result")
	}

	// Now do the same but with perms filtering.
	ctx = actor.WithActor(ctx, &actor.Actor{
		UID: 1,
	})
	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
		return authz.None, nil
	})
	authz.DefaultSubRepoPermsChecker = checker
	results, err = DefaultClient.SymbolInfo(ctx, path2)
	if err != nil {
		t.Fatalf("unexpected error when getting a definition for an unauthorized path: %s", err)
	}
	// Make sure we do not get results.
	if results != nil {
		t.Fatal("expected nil result when getting a definition for an unauthorized path")
	}
}
