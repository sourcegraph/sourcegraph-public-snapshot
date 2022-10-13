package gitserver

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

type localClient struct {
	remoteClient *clientImplementor
}

func (c *localClient) AddrForRepo(ctx context.Context, name api.RepoName) (string, error) {
	return c.remoteClient.AddrForRepo(ctx, name)
}
func (c *localClient) ArchiveReader(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, options ArchiveOptions) (io.ReadCloser, error) {
	return c.remoteClient.ArchiveReader(ctx, checker, repo, options)
}
func (c *localClient) BatchLog(ctx context.Context, opts BatchLogOptions, callback BatchLogCallback) error {
	return c.remoteClient.BatchLog(ctx, opts, callback)
}
func (c *localClient) BlameFile(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, path string, opt *BlameOptions) ([]*Hunk, error) {
	return c.remoteClient.BlameFile(ctx, checker, repo, path, opt)
}
func (c *localClient) CreateCommitFromPatch(ctx context.Context, p protocol.CreateCommitFromPatchRequest) (string, error) {
	return c.remoteClient.CreateCommitFromPatch(ctx, p)
}
func (c *localClient) GetDefaultBranch(ctx context.Context, repo api.RepoName, short bool) (refName string, commit api.CommitID, err error) {
	return c.remoteClient.GetDefaultBranch(ctx, repo, short)
}
func (c *localClient) GetObject(ctx context.Context, repo api.RepoName, objectName string) (*gitdomain.GitObject, error) {
	return c.remoteClient.GetObject(ctx, repo, objectName)
}
func (c *localClient) HasCommitAfter(ctx context.Context, repo api.RepoName, date string, revspec string, checker authz.SubRepoPermissionChecker) (bool, error) {
	return c.remoteClient.HasCommitAfter(ctx, repo, date, revspec, checker)
}
func (c *localClient) IsRepoCloneable(ctx context.Context, name api.RepoName) error {
	return c.remoteClient.IsRepoCloneable(ctx, name)
}
func (c *localClient) ListRefs(ctx context.Context, repo api.RepoName) ([]gitdomain.Ref, error) {
	return c.remoteClient.ListRefs(ctx, repo)
}
func (c *localClient) ListBranches(ctx context.Context, repo api.RepoName, opt BranchesOptions) ([]*gitdomain.Branch, error) {
	return c.remoteClient.ListBranches(ctx, repo, opt)
}
func (c *localClient) MergeBase(ctx context.Context, repo api.RepoName, a, b api.CommitID) (api.CommitID, error) {
	return c.remoteClient.MergeBase(ctx, repo, a, b)
}
func (c *localClient) P4Exec(ctx context.Context, host, user, password string, args ...string) (io.ReadCloser, http.Header, error) {
	return c.remoteClient.P4Exec(ctx, host, user, password, args...)
}
func (c *localClient) Remove(ctx context.Context, name api.RepoName) error {
	return c.remoteClient.Remove(ctx, name)
}
func (c *localClient) RemoveFrom(ctx context.Context, repo api.RepoName, from string) error {
	return c.remoteClient.RemoveFrom(ctx, repo, from)
}
func (c *localClient) RendezvousAddrForRepo(name api.RepoName) string {
	return c.remoteClient.RendezvousAddrForRepo(name)
}
func (c *localClient) RepoCloneProgress(ctx context.Context, name ...api.RepoName) (*protocol.RepoCloneProgressResponse, error) {
	return c.remoteClient.RepoCloneProgress(ctx, name...)
}
func (c *localClient) ResolveRevision(ctx context.Context, repo api.RepoName, spec string, opt ResolveRevisionOptions) (api.CommitID, error) {
	return c.remoteClient.ResolveRevision(ctx, repo, spec, opt)
}
func (c *localClient) ResolveRevisions(ctx context.Context, repo api.RepoName, specifier []protocol.RevisionSpecifier) ([]string, error) {
	return c.remoteClient.ResolveRevisions(ctx, repo, specifier)
}
func (c *localClient) ReposStats(ctx context.Context) (map[string]*protocol.ReposStats, error) {
	return c.remoteClient.ReposStats(ctx)
}
func (c *localClient) RequestRepoMigrate(ctx context.Context, repo api.RepoName, from, to string) (*protocol.RepoUpdateResponse, error) {
	return c.remoteClient.RequestRepoMigrate(ctx, repo, from, to)
}
func (c *localClient) RequestRepoUpdate(ctx context.Context, name api.RepoName, d time.Duration) (*protocol.RepoUpdateResponse, error) {
	return c.remoteClient.RequestRepoUpdate(ctx, name, d)
}

// func (c *localClient) RequestRepoClone(context.Context, api.RepoName) (*protocol.RepoCloneResponse, error) {
// 	return c.remoteClient.RequestRepoClone(context.Context, api.RepoName)
// }
// func (c *localClient) Search(_ context.Context, _ *protocol.SearchRequest, onMatches func([]protocol.CommitMatch)) (limitHit bool, _ error) {
// 	return c.remoteClient.Search(_ context.Context, _ *protocol.SearchRequest, onMatches func([]protocol.CommitMatch))
// }
// func (c *localClient) Stat(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error) {
// 	return c.remoteClient.Stat(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string)
// }
// func (c *localClient) DiffPath(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, sourceCommit, targetCommit, path string) ([]*diff.Hunk, error) {
// 	return c.remoteClient.DiffPath(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, sourceCommit, targetCommit, path string)
// }
// func (c *localClient) ReadDir(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string, recurse bool) ([]fs.FileInfo, error) {
// 	return c.remoteClient.ReadDir(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string, recurse bool)
// }
// func (c *localClient) NewFileReader(ctx context.Context, repo api.RepoName, commit api.CommitID, name string, checker authz.SubRepoPermissionChecker) (io.ReadCloser, error) {
// 	return c.remoteClient.NewFileReader(ctx context.Context, repo api.RepoName, commit api.CommitID, name string, checker authz.SubRepoPermissionChecker)
// }
// func (c *localClient) DiffSymbols(ctx context.Context, repo api.RepoName, commitA, commitB api.CommitID) ([]byte, error) {
// 	return c.remoteClient.DiffSymbols(ctx context.Context, repo api.RepoName, commitA, commitB api.CommitID)
// }
// func (c *localClient) ListFiles(ctx context.Context, repo api.RepoName, commit api.CommitID, pattern *regexp.Regexp, checker authz.SubRepoPermissionChecker) ([]string, error) {
// 	return c.remoteClient.ListFiles(ctx context.Context, repo api.RepoName, commit api.CommitID, pattern *regexp.Regexp, checker authz.SubRepoPermissionChecker)
// }
// func (c *localClient) Commits(ctx context.Context, repo api.RepoName, opt CommitsOptions, checker authz.SubRepoPermissionChecker) ([]*gitdomain.Commit, error) {
// 	return c.remoteClient.Commits(ctx context.Context, repo api.RepoName, opt CommitsOptions, checker authz.SubRepoPermissionChecker)
// }
// func (c *localClient) FirstEverCommit(ctx context.Context, repo api.RepoName, checker authz.SubRepoPermissionChecker) (*gitdomain.Commit, error) {
// 	return c.remoteClient.FirstEverCommit(ctx context.Context, repo api.RepoName, checker authz.SubRepoPermissionChecker)
// }
// func (c *localClient) ListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) ([]*gitdomain.Tag, error) {
// 	return c.remoteClient.ListTags(ctx context.Context, repo api.RepoName, commitObjs ...string)
// }
// func (c *localClient) ListDirectoryChildren(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, dirnames []string) (map[string][]string, error) {
// 	return c.remoteClient.ListDirectoryChildren(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, dirnames []string)
// }
// func (c *localClient) Diff(ctx context.Context, opts DiffOptions, checker authz.SubRepoPermissionChecker) (*DiffFileIterator, error) {
// 	return c.remoteClient.Diff(ctx context.Context, opts DiffOptions, checker authz.SubRepoPermissionChecker)
// }
// func (c *localClient) ReadFile(ctx context.Context, repo api.RepoName, commit api.CommitID, name string, checker authz.SubRepoPermissionChecker) ([]byte, error) {
// 	return c.remoteClient.ReadFile(ctx context.Context, repo api.RepoName, commit api.CommitID, name string, checker authz.SubRepoPermissionChecker)
// }
// func (c *localClient) BranchesContaining(ctx context.Context, repo api.RepoName, commit api.CommitID, checker authz.SubRepoPermissionChecker) ([]string, error) {
// 	return c.remoteClient.BranchesContaining(ctx context.Context, repo api.RepoName, commit api.CommitID, checker authz.SubRepoPermissionChecker)
// }
// func (c *localClient) RefDescriptions(ctx context.Context, repo api.RepoName, checker authz.SubRepoPermissionChecker, gitObjs ...string) (map[string][]gitdomain.RefDescription, error) {
// 	return c.remoteClient.RefDescriptions(ctx context.Context, repo api.RepoName, checker authz.SubRepoPermissionChecker, gitObjs ...string)
// }
// func (c *localClient) CommitExists(ctx context.Context, repo api.RepoName, id api.CommitID, checker authz.SubRepoPermissionChecker) (bool, error) {
// 	return c.remoteClient.CommitExists(ctx context.Context, repo api.RepoName, id api.CommitID, checker authz.SubRepoPermissionChecker)
// }
// func (c *localClient) CommitsExist(ctx context.Context, repoCommits []api.RepoCommit, checker authz.SubRepoPermissionChecker) ([]bool, error) {
// 	return c.remoteClient.CommitsExist(ctx context.Context, repoCommits []api.RepoCommit, checker authz.SubRepoPermissionChecker)
// }
// func (c *localClient) Head(ctx context.Context, repo api.RepoName, checker authz.SubRepoPermissionChecker) (string, bool, error) {
// 	return c.remoteClient.Head(ctx context.Context, repo api.RepoName, checker authz.SubRepoPermissionChecker)
// }
// func (c *localClient) CommitDate(ctx context.Context, repo api.RepoName, commit api.CommitID, checker authz.SubRepoPermissionChecker) (string, time.Time, bool, error) {
// 	return c.remoteClient.CommitDate(ctx context.Context, repo api.RepoName, commit api.CommitID, checker authz.SubRepoPermissionChecker)
// }
// func (c *localClient) CommitGraph(ctx context.Context, repo api.RepoName, opts CommitGraphOptions) (_ *gitdomain.CommitGraph, err error) {
// 	return c.remoteClient.CommitGraph(ctx context.Context, repo api.RepoName, opts CommitGraphOptions)
// }
// func (c *localClient) CommitsUniqueToBranch(ctx context.Context, repo api.RepoName, branchName string, isDefaultBranch bool, maxAge *time.Time, checker authz.SubRepoPermissionChecker) (map[string]time.Time, error) {
// 	return c.remoteClient.CommitsUniqueToBranch(ctx context.Context, repo api.RepoName, branchName string, isDefaultBranch bool, maxAge *time.Time, checker authz.SubRepoPermissionChecker)
// }
// func (c *localClient) LsFiles(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, pathspecs ...gitdomain.Pathspec) ([]string, error) {
// 	return c.remoteClient.LsFiles(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, pathspecs ...gitdomain.Pathspec)
// }
// func (c *localClient) GetCommits(ctx context.Context, repoCommits []api.RepoCommit, ignoreErrors bool, checker authz.SubRepoPermissionChecker) ([]*gitdomain.Commit, error) {
// 	return c.remoteClient.GetCommits(ctx context.Context, repoCommits []api.RepoCommit, ignoreErrors bool, checker authz.SubRepoPermissionChecker)
// }
// func (c *localClient) GetCommit(ctx context.Context, repo api.RepoName, id api.CommitID, opt ResolveRevisionOptions, checker authz.SubRepoPermissionChecker) (*gitdomain.Commit, error) {
// 	return c.remoteClient.GetCommit(ctx context.Context, repo api.RepoName, id api.CommitID, opt ResolveRevisionOptions, checker authz.SubRepoPermissionChecker)
// }
// func (c *localClient) GetBehindAhead(ctx context.Context, repo api.RepoName, left, right string) (*gitdomain.BehindAhead, error) {
// 	return c.remoteClient.GetBehindAhead(ctx context.Context, repo api.RepoName, left, right string)
// }
// func (c *localClient) ContributorCount(ctx context.Context, repo api.RepoName, opt ContributorOptions) ([]*gitdomain.ContributorCount, error) {
// 	return c.remoteClient.ContributorCount(ctx context.Context, repo api.RepoName, opt ContributorOptions)
// }
// func (c *localClient) LogReverseEach(ctx context.Context, repo string, commit string, n int, onLogEntry func(entry gitdomain.LogEntry) error) error {
// 	return c.remoteClient.LogReverseEach(ctx context.Context, repo string, commit string, n int, onLogEntry func(entry gitdomain.LogEntry) error)
// }
// func (c *localClient) RevList(ctx context.Context, repo string, commit string, onCommit func(commit string) (bool, error)) error {
// 	return c.remoteClient.RevList(ctx context.Context, repo string, commit string, onCommit func(commit string) (bool, error))
// }
// func (c *localClient) Addrs() []string {
// 	return c.remoteClient.Addrs()
// }
