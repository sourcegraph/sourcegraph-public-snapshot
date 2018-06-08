package testing

import (
	"context"
	"io"
	"os"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
)

type MockRepository struct {
	String_ func() string

	ResolveRevision_ func(ctx context.Context, spec string, opt *vcs.ResolveRevisionOptions) (api.CommitID, error)

	Branches_ func(context.Context, vcs.BranchesOptions) ([]*vcs.Branch, error)
	Tags_     func(context.Context) ([]*vcs.Tag, error)

	GetCommit_   func(context.Context, api.CommitID) (*vcs.Commit, error)
	Commits_     func(context.Context, vcs.CommitsOptions) ([]*vcs.Commit, error)
	CommitCount_ func(context.Context, vcs.CommitsOptions) (uint, error)

	BlameFile_  func(ctx context.Context, path string, opt *vcs.BlameOptions) ([]*vcs.Hunk, error)
	ExecReader_ func(ctx context.Context, args []string) (io.ReadCloser, error)
	GitCmdRaw_  func(ctx context.Context, params []string) (string, error)

	Lstat_    func(ctx context.Context, commit api.CommitID, name string) (os.FileInfo, error)
	Stat_     func(ctx context.Context, commit api.CommitID, name string) (os.FileInfo, error)
	ReadFile_ func(ctx context.Context, commit api.CommitID, name string) ([]byte, error)
	ReadDir_  func(ctx context.Context, commit api.CommitID, name string, recurse bool) ([]os.FileInfo, error)

	MergeBase_ func(ctx context.Context, a, b api.CommitID) (api.CommitID, error)

	ShortLog_ func(ctx context.Context, opt vcs.ShortLogOptions) ([]*vcs.PersonCount, error)

	Search_ func(context.Context, api.CommitID, vcs.SearchOptions) ([]*vcs.SearchResult, error)

	RawLogDiffSearch_ func(ctx context.Context, opt vcs.RawLogDiffSearchOptions) ([]*vcs.LogCommitSearchResult, bool, error)

	BehindAhead_ func(ctx context.Context, left, right string) (*vcs.BehindAhead, error)
}

var _ vcs.Repository = MockRepository{}

func (r MockRepository) String() string {
	return r.String_()
}

func (r MockRepository) ResolveRevision(ctx context.Context, spec string, opt *vcs.ResolveRevisionOptions) (api.CommitID, error) {
	return r.ResolveRevision_(ctx, spec, opt)
}

func (r MockRepository) Branches(ctx context.Context, opt vcs.BranchesOptions) ([]*vcs.Branch, error) {
	return r.Branches_(ctx, opt)
}

func (r MockRepository) Tags(ctx context.Context) ([]*vcs.Tag, error) {
	return r.Tags_(ctx)
}

func (r MockRepository) GetCommit(ctx context.Context, id api.CommitID) (*vcs.Commit, error) {
	return r.GetCommit_(ctx, id)
}

func (r MockRepository) Commits(ctx context.Context, opt vcs.CommitsOptions) ([]*vcs.Commit, error) {
	return r.Commits_(ctx, opt)
}

func (r MockRepository) CommitCount(ctx context.Context, opt vcs.CommitsOptions) (uint, error) {
	return r.CommitCount_(ctx, opt)
}

func (r MockRepository) ExecReader(ctx context.Context, args []string) (io.ReadCloser, error) {
	return r.ExecReader_(ctx, args)
}

func (r MockRepository) GitCmdRaw(ctx context.Context, params []string) (string, error) {
	return r.GitCmdRaw_(ctx, params)
}

func (r MockRepository) BlameFile(ctx context.Context, path string, opt *vcs.BlameOptions) ([]*vcs.Hunk, error) {
	return r.BlameFile_(ctx, path, opt)
}

func (r MockRepository) Lstat(ctx context.Context, commit api.CommitID, name string) (os.FileInfo, error) {
	return r.Lstat_(ctx, commit, name)
}

func (r MockRepository) Stat(ctx context.Context, commit api.CommitID, name string) (os.FileInfo, error) {
	return r.Stat_(ctx, commit, name)
}

func (r MockRepository) ReadFile(ctx context.Context, commit api.CommitID, name string) ([]byte, error) {
	return r.ReadFile_(ctx, commit, name)
}

func (r MockRepository) ReadDir(ctx context.Context, commit api.CommitID, name string, recurse bool) ([]os.FileInfo, error) {
	return r.ReadDir_(ctx, commit, name, recurse)
}

func (r MockRepository) MergeBase(ctx context.Context, a, b api.CommitID) (api.CommitID, error) {
	return r.MergeBase_(ctx, a, b)
}

func (r MockRepository) ShortLog(ctx context.Context, opt vcs.ShortLogOptions) ([]*vcs.PersonCount, error) {
	return r.ShortLog_(ctx, opt)
}

func (r MockRepository) Search(ctx context.Context, commit api.CommitID, opt vcs.SearchOptions) ([]*vcs.SearchResult, error) {
	return r.Search_(ctx, commit, opt)
}

func (r MockRepository) RawLogDiffSearch(ctx context.Context, opt vcs.RawLogDiffSearchOptions) ([]*vcs.LogCommitSearchResult, bool, error) {
	return r.RawLogDiffSearch_(ctx, opt)
}

func (r MockRepository) BehindAhead(ctx context.Context, left, right string) (*vcs.BehindAhead, error) {
	return r.BehindAhead_(ctx, left, right)
}
