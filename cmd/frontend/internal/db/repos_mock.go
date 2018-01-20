package db

import (
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
)

type MockRepos struct {
	Get      func(ctx context.Context, repo int32) (*api.Repo, error)
	GetURI   func(ctx context.Context, repo int32) (string, error)
	GetByURI func(ctx context.Context, repo string) (*api.Repo, error)
	List     func(v0 context.Context, v1 ReposListOptions) ([]*api.Repo, error)
	Delete   func(ctx context.Context, repo int32) error
	Count    func(ctx context.Context, opt ReposListOptions) (int, error)
}

func (s *MockRepos) MockGet(t *testing.T, wantRepo int32) (called *bool) {
	called = new(bool)
	s.Get = func(ctx context.Context, repo int32) (*api.Repo, error) {
		*called = true
		if repo != wantRepo {
			t.Errorf("got repo %d, want %d", repo, wantRepo)
			return nil, legacyerr.Errorf(legacyerr.NotFound, "repo %v not found", wantRepo)
		}
		return &api.Repo{ID: repo}, nil
	}
	return
}

func (s *MockRepos) MockGet_Return(t *testing.T, returns *api.Repo) (called *bool) {
	called = new(bool)
	s.Get = func(ctx context.Context, repo int32) (*api.Repo, error) {
		*called = true
		if repo != returns.ID {
			t.Errorf("got repo %d, want %d", repo, returns.ID)
			return nil, legacyerr.Errorf(legacyerr.NotFound, "repo %v (%d) not found", returns.URI, returns.ID)
		}
		return returns, nil
	}
	return
}

func (s *MockRepos) MockGetURI(t *testing.T, want int32, returns string) (called *bool) {
	called = new(bool)
	s.GetURI = func(ctx context.Context, repo int32) (string, error) {
		*called = true
		if repo != want {
			t.Errorf("got repo %d, want %d", repo, want)
			return "", legacyerr.Errorf(legacyerr.NotFound, "repo %d not found", want)
		}
		return returns, nil
	}
	return
}

func (s *MockRepos) MockGetByURI(t *testing.T, wantURI string, repoID int32) (called *bool) {
	called = new(bool)
	s.GetByURI = func(ctx context.Context, uri string) (*api.Repo, error) {
		*called = true
		if uri != wantURI {
			t.Errorf("got repo URI %q, want %q", uri, wantURI)
			return nil, legacyerr.Errorf(legacyerr.NotFound, "repo %v not found", uri)
		}
		return &api.Repo{ID: repoID, URI: uri}, nil
	}
	return
}

func (s *MockRepos) MockList(t *testing.T, wantRepos ...string) (called *bool) {
	called = new(bool)
	s.List = func(ctx context.Context, opt ReposListOptions) ([]*api.Repo, error) {
		*called = true
		repos := make([]*api.Repo, len(wantRepos))
		for i, repo := range wantRepos {
			repos[i] = &api.Repo{URI: repo}
		}
		return repos, nil
	}
	return
}
