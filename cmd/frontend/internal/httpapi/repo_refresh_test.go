package httpapi

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRepoRefresh(t *testing.T) {
	c := newTest()

	enqueueRepoUpdateCount := map[api.RepoName]int{}
	gitserver.MockRequestRepoUpdate = func(ctx context.Context, repo api.RepoName, duration time.Duration) (*protocol.RepoUpdateResponse, error) {
		enqueueRepoUpdateCount[repo]++
		return nil, nil
	}
	t.Cleanup(func() {
		gitserver.MockRequestRepoUpdate = nil
	})

	backend.Mocks.Repos.GetByName = func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
		switch name {
		case "github.com/gorilla/mux":
			return &types.Repo{ID: 2, Name: name}, nil
		default:
			panic("wrong path")
		}
	}
	t.Cleanup(func() {
		backend.Mocks.Repos.GetByName = nil
	})

	if _, err := c.PostOK("/repos/github.com/gorilla/mux/-/refresh", nil); err != nil {
		t.Fatal(err)
	}
	if ct := enqueueRepoUpdateCount["github.com/gorilla/mux"]; ct != 1 {
		t.Errorf("expected EnqueueRepoUpdate to be called once, but was called %d times", ct)
	}
}
