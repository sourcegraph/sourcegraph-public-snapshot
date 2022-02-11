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
)

func TestSearchWithFiltering(t *testing.T) {
	ctx := context.Background()
	fixture := result.Symbols{
		result.Symbol{
			Name: "foo1",
			Path: "file1",
		},
		result.Symbol{
			Name: "foo2",
			Path: "file2",
		},
	}
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
