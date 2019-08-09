package backend

import (
	"sync"
	"testing"

	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

type MockRepos struct {
	Get                       func(v0 context.Context, id api.RepoID) (*types.Repo, error)
	GetByName                 func(v0 context.Context, name api.RepoName) (*types.Repo, error)
	AddGitHubDotComRepository func(name api.RepoName) error
	List                      func(v0 context.Context, v1 db.ReposListOptions) ([]*types.Repo, error)
	GetCommit                 func(v0 context.Context, repo *types.Repo, commitID api.CommitID) (*git.Commit, error)
	ResolveRev                func(v0 context.Context, repo *types.Repo, rev string) (api.CommitID, error)
	GetInventory              func(v0 context.Context, repo *types.Repo, commitID api.CommitID) (*inventory.Inventory, error)
	GetInventoryUncached      func(ctx context.Context, repo *types.Repo, commitID api.CommitID) (*inventory.Inventory, error)
}

var errRepoNotFound = &errcode.Mock{
	Message:    "repo not found",
	IsNotFound: true,
}

func (s *MockRepos) MockGet(t *testing.T, wantRepo api.RepoID) (called *bool) {
	called = new(bool)
	s.Get = func(ctx context.Context, repo api.RepoID) (*types.Repo, error) {
		*called = true
		if repo != wantRepo {
			t.Errorf("got repo %d, want %d", repo, wantRepo)
			return nil, errRepoNotFound
		}
		return &types.Repo{ID: repo}, nil
	}
	return
}

func (s *MockRepos) MockGetByName(t *testing.T, wantName api.RepoName, repo api.RepoID) (called *bool) {
	called = new(bool)
	s.GetByName = func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
		*called = true
		if name != wantName {
			t.Errorf("got repo name %q, want %q", name, wantName)
			return nil, errRepoNotFound
		}
		return &types.Repo{ID: repo, Name: name}, nil
	}
	return
}

func (s *MockRepos) MockGet_Return(t *testing.T, returns *types.Repo) (called *bool) {
	called = new(bool)
	s.Get = func(ctx context.Context, repo api.RepoID) (*types.Repo, error) {
		*called = true
		if repo != returns.ID {
			t.Errorf("got repo %d, want %d", repo, returns.ID)
			return nil, errRepoNotFound
		}
		return returns, nil
	}
	return
}

func (s *MockRepos) MockList(t *testing.T, wantRepos ...api.RepoName) (called *bool) {
	called = new(bool)
	s.List = func(ctx context.Context, opt db.ReposListOptions) ([]*types.Repo, error) {
		*called = true
		repos := make([]*types.Repo, len(wantRepos))
		for i, repo := range wantRepos {
			repos[i] = &types.Repo{Name: repo}
		}
		return repos, nil
	}
	return
}

func (s *MockRepos) MockResolveRev_NoCheck(t *testing.T, commitID api.CommitID) (called *bool) {
	var once sync.Once
	called = new(bool)
	s.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		once.Do(func() {
			*called = true
		})
		return commitID, nil
	}
	return
}

func (s *MockRepos) MockResolveRev_NotFound(t *testing.T, wantRepo api.RepoID, wantRev string) (called *bool) {
	called = new(bool)
	s.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		*called = true
		if repo.ID != wantRepo {
			t.Errorf("got repo %v, want %v", repo.ID, wantRepo)
		}
		if rev != wantRev {
			t.Errorf("got rev %v, want %v", rev, wantRev)
		}
		return "", &gitserver.RevisionNotFoundError{Repo: repo.Name, Spec: rev}
	}
	return
}

func (s *MockRepos) MockGetCommit_Return_NoCheck(t *testing.T, commit *git.Commit) (called *bool) {
	called = new(bool)
	s.GetCommit = func(ctx context.Context, repo *types.Repo, commitID api.CommitID) (*git.Commit, error) {
		*called = true
		return commit, nil
	}
	return
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_24(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
