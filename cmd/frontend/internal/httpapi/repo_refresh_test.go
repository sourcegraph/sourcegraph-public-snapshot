package httpapi

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
)

func TestRepoRefresh(t *testing.T) {
	c := newTest()

	enqueueRepoUpdateCount := map[api.RepoURI]int{}
	repoupdater.MockEnqueueRepoUpdate = func(ctx context.Context, repo gitserver.Repo) error {
		enqueueRepoUpdateCount[repo.Name]++
		return nil
	}

	backend.Mocks.Repos.GetByURI = func(ctx context.Context, uri api.RepoURI) (*types.Repo, error) {
		switch uri {
		case "github.com/gorilla/mux":
			return &types.Repo{ID: 2, URI: uri}, nil
		default:
			panic("wrong path")
		}
	}
	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		if repo.ID != 2 || rev != "master" {
			t.Error("wrong arguments to ResolveRev")
		}
		return "aed", nil
	}

	if _, err := c.PostOK("/repos/github.com/gorilla/mux/-/refresh", nil); err != nil {
		t.Fatal(err)
	}
	if ct := enqueueRepoUpdateCount["github.com/gorilla/mux"]; ct != 1 {
		t.Errorf("expected EnqueueRepoUpdate to be called once, but was called %d times", ct)
	}
}
