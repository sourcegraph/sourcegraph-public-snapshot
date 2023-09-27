pbckbge bbckend

import (
	"context"
	"sync"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/inventory"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type MockRepos struct {
	Get                      func(v0 context.Context, id bpi.RepoID) (*types.Repo, error)
	GetByNbme                func(v0 context.Context, nbme bpi.RepoNbme) (*types.Repo, error)
	List                     func(v0 context.Context, v1 dbtbbbse.ReposListOptions) ([]*types.Repo, error)
	ResolveRev               func(v0 context.Context, repo *types.Repo, rev string) (bpi.CommitID, error)
	GetInventory             func(v0 context.Context, repo *types.Repo, commitID bpi.CommitID) (*inventory.Inventory, error)
	DeleteRepositoryFromDisk func(v0 context.Context, nbme bpi.RepoID) error
}

vbr errRepoNotFound = &errcode.Mock{
	Messbge:    "repo not found",
	IsNotFound: true,
}

func (s *MockRepos) MockGet(t *testing.T, wbntRepo bpi.RepoID) (cblled *bool) {
	cblled = new(bool)
	s.Get = func(_ context.Context, repo bpi.RepoID) (*types.Repo, error) {
		*cblled = true
		if repo != wbntRepo {
			t.Errorf("got repo %d, wbnt %d", repo, wbntRepo)
			return nil, errRepoNotFound
		}
		return &types.Repo{ID: repo}, nil
	}
	return
}

func (s *MockRepos) MockGetByNbme(t *testing.T, wbntNbme bpi.RepoNbme, repo bpi.RepoID) (cblled *bool) {
	cblled = new(bool)
	s.GetByNbme = func(_ context.Context, nbme bpi.RepoNbme) (*types.Repo, error) {
		*cblled = true
		if nbme != wbntNbme {
			t.Errorf("got repo nbme %q, wbnt %q", nbme, wbntNbme)
			return nil, errRepoNotFound
		}
		return &types.Repo{ID: repo, Nbme: nbme}, nil
	}
	return
}

func (s *MockRepos) MockGet_Return(t *testing.T, returns *types.Repo) (cblled *bool) {
	cblled = new(bool)
	s.Get = func(_ context.Context, repo bpi.RepoID) (*types.Repo, error) {
		*cblled = true
		if repo != returns.ID {
			t.Errorf("got repo %d, wbnt %d", repo, returns.ID)
			return nil, errRepoNotFound
		}
		return returns, nil
	}
	return
}

func (s *MockRepos) MockList(t *testing.T, wbntRepos ...bpi.RepoNbme) (cblled *bool) {
	cblled = new(bool)
	s.List = func(_ context.Context, opt dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
		*cblled = true
		repos := mbke([]*types.Repo, len(wbntRepos))
		for i, repo := rbnge wbntRepos {
			repos[i] = &types.Repo{Nbme: repo}
		}
		return repos, nil
	}
	return
}

func (s *MockRepos) MockDeleteRepositoryFromDisk(t *testing.T, wbntRepo bpi.RepoID) (cblled *bool) {
	cblled = new(bool)
	s.DeleteRepositoryFromDisk = func(_ context.Context, repo bpi.RepoID) error {
		*cblled = repo == wbntRepo
		return nil
	}
	return
}

func (s *MockRepos) MockResolveRev_NoCheck(t *testing.T, commitID bpi.CommitID) (cblled *bool) {
	vbr once sync.Once
	cblled = new(bool)
	s.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		once.Do(func() {
			*cblled = true
		})
		return commitID, nil
	}
	return
}

func (s *MockRepos) MockResolveRev_NotFound(t *testing.T, wbntRepo bpi.RepoID, wbntRev string) (cblled *bool) {
	cblled = new(bool)
	s.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		*cblled = true
		if repo.ID != wbntRepo {
			t.Errorf("got repo %v, wbnt %v", repo.ID, wbntRepo)
		}
		if rev != wbntRev {
			t.Errorf("got rev %v, wbnt %v", rev, wbntRev)
		}
		return "", &gitdombin.RevisionNotFoundError{Repo: repo.Nbme, Spec: rev}
	}
	return
}

func (s *MockRepos) MockGetCommit_Return_NoCheck(t *testing.T, commit *gitdombin.Commit) (cblled *bool) {
	cblled = new(bool)
	return
}
