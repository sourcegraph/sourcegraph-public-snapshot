package idx

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/graphqlbackend"
)

// svc abstracts all the other services the indexer depends on for
// data. It allows these services to be easily mocked.
type svc interface {
	ListPackages(ctx context.Context, op *sourcegraph.ListPackagesOp) (pkgs []sourcegraph.PackageInfo, err error)
	GetInventoryUncached(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*inventory.Inventory, error)
	Update(ctx context.Context, op *sourcegraph.ReposUpdateOp) (err error)
	GetByURI(ctx context.Context, uri string) (res *sourcegraph.Repo, err error)
	Dependencies(ctx context.Context, repoID int32, excludePrivate bool) ([]*sourcegraph.DependencyReference, error)
	ResolveRepo(ctx context.Context, uri string) (*sourcegraph.Repo, error)
	ResolveRevision(ctx context.Context, repo *sourcegraph.Repo, spec string) (vcs.CommitID, error)
	DefsRefreshIndex(ctx context.Context, repoURI, commit string) (err error)
	PkgsRefreshIndex(ctx context.Context, repo string, commit string) (err error)
	GoogleGitHub(query string) (string, error)
}

var DefaultSvc = &svcImpl{}

type svcImpl struct{}

func (s *svcImpl) ListPackages(ctx context.Context, op *sourcegraph.ListPackagesOp) (pkgs []sourcegraph.PackageInfo, err error) {
	return backend.Pkgs.ListPackages(ctx, op)
}
func (s *svcImpl) GetInventoryUncached(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*inventory.Inventory, error) {
	return backend.Repos.GetInventoryUncached(ctx, repoRev)
}
func (s *svcImpl) Update(ctx context.Context, op *sourcegraph.ReposUpdateOp) (err error) {
	return backend.Repos.Update(ctx, op)
}
func (s *svcImpl) GetByURI(ctx context.Context, uri string) (res *sourcegraph.Repo, err error) {
	return backend.Repos.GetByURI(ctx, uri)
}
func (s *svcImpl) Dependencies(ctx context.Context, repoID int32, excludePrivate bool) ([]*sourcegraph.DependencyReference, error) {
	return backend.Defs.Dependencies(ctx, repoID, excludePrivate)
}
func (s *svcImpl) ResolveRepo(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	return graphqlbackend.ResolveRepo(ctx, uri)
}
func (s *svcImpl) ResolveRevision(ctx context.Context, repo *sourcegraph.Repo, spec string) (vcs.CommitID, error) {
	return gitcmd.Open(repo).ResolveRevision(ctx, spec)
}
func (s *svcImpl) DefsRefreshIndex(ctx context.Context, repoURI, commit string) (err error) {
	return backend.Defs.RefreshIndex(ctx, repoURI, commit)
}
func (s *svcImpl) PkgsRefreshIndex(ctx context.Context, repo string, commit string) (err error) {
	return backend.Pkgs.RefreshIndex(ctx, repo, commit)
}
func (s *svcImpl) GoogleGitHub(query string) (string, error) {
	return Google.Search(query)
}

type svcMock struct {
	ListPackages_         func(ctx context.Context, op *sourcegraph.ListPackagesOp) (pkgs []sourcegraph.PackageInfo, err error)
	GetInventoryUncached_ func(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*inventory.Inventory, error)
	Update_               func(ctx context.Context, op *sourcegraph.ReposUpdateOp) (err error)
	GetByURI_             func(ctx context.Context, uri string) (res *sourcegraph.Repo, err error)
	Dependencies_         func(ctx context.Context, repoID int32, excludePrivate bool) ([]*sourcegraph.DependencyReference, error)
	ResolveRepo_          func(ctx context.Context, uri string) (*sourcegraph.Repo, error)
	ResolveRevision_      func(ctx context.Context, repo *sourcegraph.Repo, spec string) (vcs.CommitID, error)
	DefsRefreshIndex_     func(ctx context.Context, repoURI, commit string) (err error)
	PkgsRefreshIndex_     func(ctx context.Context, repo string, commit string) (err error)
	GoogleGitHub_         func(query string) (string, error)
}

func (s svcMock) ListPackages(ctx context.Context, op *sourcegraph.ListPackagesOp) (pkgs []sourcegraph.PackageInfo, err error) {
	return s.ListPackages_(ctx, op)
}
func (s svcMock) GetInventoryUncached(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*inventory.Inventory, error) {
	return s.GetInventoryUncached_(ctx, repoRev)
}
func (s svcMock) Update(ctx context.Context, op *sourcegraph.ReposUpdateOp) (err error) {
	return s.Update_(ctx, op)
}
func (s svcMock) GetByURI(ctx context.Context, uri string) (res *sourcegraph.Repo, err error) {
	return s.GetByURI_(ctx, uri)
}
func (s svcMock) Dependencies(ctx context.Context, repoID int32, excludePrivate bool) ([]*sourcegraph.DependencyReference, error) {
	return s.Dependencies_(ctx, repoID, excludePrivate)
}
func (s svcMock) ResolveRepo(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	return s.ResolveRepo_(ctx, uri)
}
func (s svcMock) ResolveRevision(ctx context.Context, repo *sourcegraph.Repo, spec string) (vcs.CommitID, error) {
	return s.ResolveRevision_(ctx, repo, spec)
}
func (s svcMock) DefsRefreshIndex(ctx context.Context, repoURI, commit string) (err error) {
	return s.DefsRefreshIndex_(ctx, repoURI, commit)
}
func (s svcMock) PkgsRefreshIndex(ctx context.Context, repo string, commit string) (err error) {
	return s.PkgsRefreshIndex_(ctx, repo, commit)
}
func (s svcMock) GoogleGitHub(query string) (string, error) {
	return s.GoogleGitHub_(query)
}
