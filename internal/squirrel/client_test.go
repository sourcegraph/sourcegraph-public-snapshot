package squirrel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSearchWithFiltering(t *testing.T) {
	ctx := context.Background()
	fixture := types.SquirrelLocation{
		Repo:   "somerepo",
		Commit: "somecommit",
		Path:   "path1",
		Row:    0,
		Column: 0,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(fixture)
	}))
	t.Cleanup(func() {
		srv.Close()
	})
	DefaultClient.URL = srv.URL

	results, err := DefaultClient.Definition(ctx, types.SquirrelLocation{
		Repo:   "somrepo",
		Commit: "somecommit",
		Path:   "path2",
		Row:    0,
		Column: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if results == nil {
		t.Fatal("nil result")
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
		return authz.None, nil
	})
	authz.DefaultSubRepoPermsChecker = checker

	results, err = DefaultClient.Definition(ctx, types.SquirrelLocation{
		Repo:   "somrepo",
		Commit: "somecommit",
		Path:   "path2",
		Row:    0,
		Column: 0,
	})
	if err == nil {
		t.Fatal("expected error when getting a definition for an unauthorized path")
	}
	if results != nil {
		t.Fatal("expected nil result when getting a definition for an unauthorized path")
	}
}
