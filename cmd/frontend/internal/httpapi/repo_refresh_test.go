package httpapi

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRepoRefresh(t *testing.T) {
	c := newTest(t)

	enqueueRepoUpdateCount := map[api.RepoName]int{}
	repoupdater.MockEnqueueRepoUpdate = func(ctx context.Context, repo api.RepoName) (*protocol.RepoUpdateResponse, error) {
		enqueueRepoUpdateCount[repo]++
		return nil, nil
	}

	backend.Mocks.Repos.GetByName = func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
		switch name {
		case "github.com/gorilla/mux":
			return &types.Repo{ID: 2, Name: name}, nil
		default:
			panic("wrong path")
		}
	}
	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo api.RepoName, rev string) (api.CommitID, error) {
		if repo != "github.com/gorilla/mux" || rev != "master" {
			t.Error("wrong arguments to ResolveRev")
		}
		return "aed", nil
	}

	if _, err := c.PostOK("/.api/repos/github.com/gorilla/mux/-/refresh", nil); err != nil {
		t.Fatal(err)
	}
	if ct := enqueueRepoUpdateCount["github.com/gorilla/mux"]; ct != 1 {
		t.Errorf("expected EnqueueRepoUpdate to be called once, but was called %d times", ct)
	}
}
