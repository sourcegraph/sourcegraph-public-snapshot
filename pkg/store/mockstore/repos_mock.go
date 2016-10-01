package mockstore

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type Repos struct {
	Get            func(ctx context.Context, repo int32) (*sourcegraph.Repo, error)
	GetByURI       func(ctx context.Context, repo string) (*sourcegraph.Repo, error)
	List           func(v0 context.Context, v1 *store.RepoListOp) ([]*sourcegraph.Repo, error)
	Search         func(v0 context.Context, v1 string) ([]*sourcegraph.RepoSearchResult, error)
	Create         func(v0 context.Context, v1 *sourcegraph.Repo) (int32, error)
	Update         func(v0 context.Context, v1 store.RepoUpdate) error
	InternalUpdate func(ctx context.Context, repo int32, op store.InternalRepoUpdate) error
	Delete         func(ctx context.Context, repo int32) error
}

type RepoConfigs struct {
	Get    func(ctx context.Context, repo int32) (*sourcegraph.RepoConfig, error)
	Update func(ctx context.Context, repo int32, conf sourcegraph.RepoConfig) error
}

type RepoStatuses struct {
	GetCombined func(ctx context.Context, repo int32, commitID string) (*sourcegraph.CombinedStatus, error)
	GetCoverage func(ctx context.Context) (*sourcegraph.RepoStatusList, error)
	Create      func(ctx context.Context, repo int32, commitID string, status *sourcegraph.RepoStatus) error
}

type RepoVCS struct {
	Open  func(ctx context.Context, repo int32) (vcs.Repository, error)
	Clone func(ctx context.Context, repo int32, info *store.CloneInfo) error
}
