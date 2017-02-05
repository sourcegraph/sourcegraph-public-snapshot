package httpapi

import (
	"context"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

func TestRepoShield(t *testing.T) {
	c := newTest()

	wantResp := map[string]interface{}{
		"value": " 200 projects",
	}

	backend.Mocks.Repos.Get = func(ctx context.Context, r *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		return &sourcegraph.Repo{
			ID:            2,
			URI:           "github.com/gorilla/mux",
			Description:   "desc",
			DefaultBranch: "master",
		}, nil
	}
	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, op *sourcegraph.ReposResolveRevOp) (*sourcegraph.ResolvedRev, error) {
		if op.Repo != 2 || op.Rev != "master" {
			t.Error("wrong arguments to ResolveRev")
		}
		return &sourcegraph.ResolvedRev{
			CommitID: "aed",
		}, nil
	}
	backend.Mocks.Defs.TotalRefs = func(ctx context.Context, source string) (int, error) {
		if source != "github.com/gorilla/mux" {
			t.Error("wrong repo source to TotalRefs")
		}
		return 200, nil
	}

	var resp map[string]interface{}
	if err := c.GetJSON("/repos/github.com/gorilla/mux/-/shield", &resp); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(resp, wantResp) {
		t.Errorf("got %+v, want %+v", resp, wantResp)
	}
}
