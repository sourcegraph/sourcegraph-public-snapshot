package testing

import (
	"context"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type MockRepository struct {
	String_ func() string

	ResolveRevision_ func(ctx context.Context, spec string) (vcs.CommitID, error)

	Branches_ func(context.Context, vcs.BranchesOptions) ([]*vcs.Branch, error)
	Tags_     func(context.Context) ([]*vcs.Tag, error)

	GetCommit_ func(context.Context, vcs.CommitID) (*vcs.Commit, error)
	Commits_   func(context.Context, vcs.CommitsOptions) ([]*vcs.Commit, uint, error)

	BlameFile_ func(ctx context.Context, path string, opt *vcs.BlameOptions) ([]*vcs.Hunk, error)

	Lstat_    func(ctx context.Context, commit vcs.CommitID, name string) (os.FileInfo, error)
	Stat_     func(ctx context.Context, commit vcs.CommitID, name string) (os.FileInfo, error)
	ReadFile_ func(ctx context.Context, commit vcs.CommitID, name string) ([]byte, error)
	ReadDir_  func(ctx context.Context, commit vcs.CommitID, name string, recurse bool) ([]os.FileInfo, error)

	Diff_      func(ctx context.Context, base, head vcs.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error)
	MergeBase_ func(ctx context.Context, a, b vcs.CommitID) (vcs.CommitID, error)

	Committers_ func(ctx context.Context, opt vcs.CommittersOptions) ([]*vcs.Committer, error)

	UpdateEverything_ func(context.Context, vcs.RemoteOpts) (*vcs.UpdateResult, error)

	Search_ func(context.Context, vcs.CommitID, vcs.SearchOptions) ([]*vcs.SearchResult, error)
}

var _ vcs.Repository = MockRepository{}

func (r MockRepository) String() string {
	return r.String_()
}

func (r MockRepository) ResolveRevision(ctx context.Context, spec string) (vcs.CommitID, error) {
	return r.ResolveRevision_(ctx, spec)
}

func (r MockRepository) Branches(ctx context.Context, opt vcs.BranchesOptions) ([]*vcs.Branch, error) {
	return r.Branches_(ctx, opt)
}

func (r MockRepository) Tags(ctx context.Context) ([]*vcs.Tag, error) {
	return r.Tags_(ctx)
}

func (r MockRepository) GetCommit(ctx context.Context, id vcs.CommitID) (*vcs.Commit, error) {
	return r.GetCommit_(ctx, id)
}

func (r MockRepository) Commits(ctx context.Context, opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error) {
	return r.Commits_(ctx, opt)
}

func (r MockRepository) BlameFile(ctx context.Context, path string, opt *vcs.BlameOptions) ([]*vcs.Hunk, error) {
	return r.BlameFile_(ctx, path, opt)
}

func (r MockRepository) Lstat(ctx context.Context, commit vcs.CommitID, name string) (os.FileInfo, error) {
	return r.Lstat_(ctx, commit, name)
}

func (r MockRepository) Stat(ctx context.Context, commit vcs.CommitID, name string) (os.FileInfo, error) {
	return r.Stat_(ctx, commit, name)
}

func (r MockRepository) ReadFile(ctx context.Context, commit vcs.CommitID, name string) ([]byte, error) {
	return r.ReadFile_(ctx, commit, name)
}

func (r MockRepository) ReadDir(ctx context.Context, commit vcs.CommitID, name string, recurse bool) ([]os.FileInfo, error) {
	return r.ReadDir_(ctx, commit, name, recurse)
}

func (r MockRepository) Diff(ctx context.Context, base, head vcs.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error) {
	return r.Diff_(ctx, base, head, opt)
}

func (r MockRepository) MergeBase(ctx context.Context, a, b vcs.CommitID) (vcs.CommitID, error) {
	return r.MergeBase_(ctx, a, b)
}

func (r MockRepository) Committers(ctx context.Context, opt vcs.CommittersOptions) ([]*vcs.Committer, error) {
	return r.Committers_(ctx, opt)
}

func (r MockRepository) UpdateEverything(ctx context.Context, opt vcs.RemoteOpts) (*vcs.UpdateResult, error) {
	return r.UpdateEverything_(ctx, opt)
}

func (r MockRepository) Search(ctx context.Context, commit vcs.CommitID, opt vcs.SearchOptions) ([]*vcs.SearchResult, error) {
	return r.Search_(ctx, commit, opt)
}
