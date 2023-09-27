pbckbge httpbpi

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestRepoRefresh(t *testing.T) {
	c := newTest(t)

	enqueueRepoUpdbteCount := mbp[bpi.RepoNbme]int{}
	repoupdbter.MockEnqueueRepoUpdbte = func(ctx context.Context, repo bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error) {
		enqueueRepoUpdbteCount[repo]++
		return nil, nil
	}

	bbckend.Mocks.Repos.GetByNbme = func(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error) {
		switch nbme {
		cbse "github.com/gorillb/mux":
			return &types.Repo{ID: 2, Nbme: nbme}, nil
		defbult:
			pbnic("wrong pbth")
		}
	}
	bbckend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		if repo.ID != 2 || rev != "mbster" {
			t.Error("wrong brguments to ResolveRev")
		}
		return "bed", nil
	}

	if _, err := c.PostOK("/repos/github.com/gorillb/mux/-/refresh", nil); err != nil {
		t.Fbtbl(err)
	}
	if ct := enqueueRepoUpdbteCount["github.com/gorillb/mux"]; ct != 1 {
		t.Errorf("expected EnqueueRepoUpdbte to be cblled once, but wbs cblled %d times", ct)
	}
}
