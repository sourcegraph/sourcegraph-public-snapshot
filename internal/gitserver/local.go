package gitserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"os/exec"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

const root = "/home/beyang/src/github.com"

type localClient struct {
	remoteClient *clientImplementor
}

func (c *localClient) AddrForRepo(ctx context.Context, name api.RepoName) (string, error) {
	fmt.Println("### localClient.AddrForRepo")
	return c.remoteClient.AddrForRepo(ctx, name)
}
func (c *localClient) ArchiveReader(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, options ArchiveOptions) (io.ReadCloser, error) {
	fmt.Println("### localClient.ArchiveReader")
	return c.remoteClient.ArchiveReader(ctx, checker, repo, options)
}
func (c *localClient) BatchLog(ctx context.Context, opts BatchLogOptions, callback BatchLogCallback) error {
	fmt.Println("### localClient.BatchLog")
	return c.remoteClient.BatchLog(ctx, opts, callback)
}
func (c *localClient) BlameFile(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, path string, opt *BlameOptions) ([]*Hunk, error) {
	fmt.Println("### localClient.BlameFile")
	return c.remoteClient.BlameFile(ctx, checker, repo, path, opt)
}
func (c *localClient) CreateCommitFromPatch(ctx context.Context, p protocol.CreateCommitFromPatchRequest) (string, error) {
	fmt.Println("### localClient.CreateCommitFromPatch")
	return c.remoteClient.CreateCommitFromPatch(ctx, p)
}
func (c *localClient) GetDefaultBranch(ctx context.Context, repo api.RepoName, short bool) (refName string, commit api.CommitID, err error) {
	if strings.HasPrefix(string(repo), "local/") {
		args := []string{"symbolic-ref", "HEAD"}
		if short {
			args = append(args, "--short")
		}
		cmd := exec.Command("git", args...)
		cmd.Dir = filepath.Join(root, strings.TrimPrefix(string(repo), "local/"))

		refBytes, err := cmd.Output()
		if err != nil {
			return "", "", err
		}
		refName = string(bytes.TrimSpace(refBytes))
		return refName, commit, nil
	}
	return c.remoteClient.GetDefaultBranch(ctx, repo, short)
}
func (c *localClient) GetObject(ctx context.Context, repo api.RepoName, objectName string) (*gitdomain.GitObject, error) {
	fmt.Println("### localClient.GetObject")
	return c.remoteClient.GetObject(ctx, repo, objectName)
}
func (c *localClient) HasCommitAfter(ctx context.Context, repo api.RepoName, date string, revspec string, checker authz.SubRepoPermissionChecker) (bool, error) {
	fmt.Println("### localClient.HasCommitAfter")
	return c.remoteClient.HasCommitAfter(ctx, repo, date, revspec, checker)
}
func (c *localClient) IsRepoCloneable(ctx context.Context, name api.RepoName) error {
	fmt.Println("### localClient.IsRepoCloneable")
	return c.remoteClient.IsRepoCloneable(ctx, name)
}
func (c *localClient) ListRefs(ctx context.Context, repo api.RepoName) ([]gitdomain.Ref, error) {
	fmt.Println("### localClient.ListRefs")
	return c.remoteClient.ListRefs(ctx, repo)
}
func (c *localClient) ListBranches(ctx context.Context, repo api.RepoName, opt BranchesOptions) ([]*gitdomain.Branch, error) {
	fmt.Println("### localClient.ListBranches")
	return c.remoteClient.ListBranches(ctx, repo, opt)
}
func (c *localClient) MergeBase(ctx context.Context, repo api.RepoName, a, b api.CommitID) (api.CommitID, error) {
	fmt.Println("### localClient.MergeBase")
	return c.remoteClient.MergeBase(ctx, repo, a, b)
}
func (c *localClient) P4Exec(ctx context.Context, host, user, password string, args ...string) (io.ReadCloser, http.Header, error) {
	fmt.Println("### localClient.P4Exec")
	return c.remoteClient.P4Exec(ctx, host, user, password, args...)
}
func (c *localClient) Remove(ctx context.Context, name api.RepoName) error {
	fmt.Println("### localClient.Remove")
	return c.remoteClient.Remove(ctx, name)
}
func (c *localClient) RemoveFrom(ctx context.Context, repo api.RepoName, from string) error {
	fmt.Println("### localClient.RemoveFrom")
	return c.remoteClient.RemoveFrom(ctx, repo, from)
}
func (c *localClient) RendezvousAddrForRepo(name api.RepoName) string {
	fmt.Println("### localClient.RendezvousAddrForRepo")
	return c.remoteClient.RendezvousAddrForRepo(name)
}
func (c *localClient) RepoCloneProgress(ctx context.Context, name ...api.RepoName) (*protocol.RepoCloneProgressResponse, error) {
	r, err := c.remoteClient.RepoCloneProgress(ctx, name...)
	if err != nil {
		return nil, err
	}
	if r.Results == nil {
		r.Results = make(map[api.RepoName]*protocol.RepoCloneProgress)
	}
	for _, rname := range name {
		if strings.HasPrefix(string(rname), "local/") {
			r.Results[rname] = &protocol.RepoCloneProgress{
				Cloned: true,
			}
		}
	}
	return r, nil
}
func (c *localClient) ResolveRevision(ctx context.Context, repo api.RepoName, spec string, opt ResolveRevisionOptions) (api.CommitID, error) {
	if strings.HasPrefix(string(repo), "local/") {
		if spec == "" {
			spec = "HEAD"
		}
		cmd := exec.Command("git", "rev-parse", spec)
		cmd.Dir = filepath.Join(root, strings.TrimPrefix(string(repo), "local/"))
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		c := strings.TrimSpace(string(out))
		if c == "" {
			return "", fmt.Errorf("empty revision value from running `git rev-parse %s`", spec)
		}
		return api.CommitID(c), nil
	}
	return c.remoteClient.ResolveRevision(ctx, repo, spec, opt)
}
func (c *localClient) ResolveRevisions(ctx context.Context, repo api.RepoName, specifier []protocol.RevisionSpecifier) ([]string, error) {
	if strings.HasPrefix(string(repo), "local/") {
		panic("ResolveRevisions")
	}
	return c.remoteClient.ResolveRevisions(ctx, repo, specifier)
}
func (c *localClient) ReposStats(ctx context.Context) (map[string]*protocol.ReposStats, error) {
	return c.remoteClient.ReposStats(ctx)
}
func (c *localClient) RequestRepoMigrate(ctx context.Context, repo api.RepoName, from, to string) (*protocol.RepoUpdateResponse, error) {
	fmt.Println("### localClient.RequestRepoMigrate")
	return c.remoteClient.RequestRepoMigrate(ctx, repo, from, to)
}
func (c *localClient) RequestRepoUpdate(ctx context.Context, name api.RepoName, d time.Duration) (*protocol.RepoUpdateResponse, error) {
	if strings.HasPrefix(string(name), "local/") {
		return nil, nil
	}
	return c.remoteClient.RequestRepoUpdate(ctx, name, d)
}
func (c *localClient) RequestRepoClone(ctx context.Context, name api.RepoName) (*protocol.RepoCloneResponse, error) {
	fmt.Println("### localClient.RequestRepoClone")
	return c.remoteClient.RequestRepoClone(ctx, name)
}
func (c *localClient) Search(ctx context.Context, sr *protocol.SearchRequest, onMatches func([]protocol.CommitMatch)) (limitHit bool, _ error) {
	fmt.Println("### localClient.Search")
	return c.remoteClient.Search(ctx, sr, onMatches)
}
func (c *localClient) Stat(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error) {
	if strings.HasPrefix(string(repo), "local/") {
		// TODO: handle non-HEAD commits
		return os.Stat(filepath.Join(root, strings.TrimPrefix(string(repo), "local/"), path))
	}
	return c.remoteClient.Stat(ctx, checker, repo, commit, path)
}
func (c *localClient) DiffPath(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, sourceCommit, targetCommit, path string) ([]*diff.Hunk, error) {
	fmt.Println("### localClient.DiffPath")
	return c.remoteClient.DiffPath(ctx, checker, repo, sourceCommit, targetCommit, path)
}
func (c *localClient) ReadDir(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string, recurse bool) ([]fs.FileInfo, error) {
	fmt.Println("### localClient.ReadDir")
	return c.remoteClient.ReadDir(ctx, checker, repo, commit, path, recurse)
}
func (c *localClient) NewFileReader(ctx context.Context, repo api.RepoName, commit api.CommitID, name string, checker authz.SubRepoPermissionChecker) (io.ReadCloser, error) {
	fmt.Println("### localClient.NewFileReader")
	return c.remoteClient.NewFileReader(ctx, repo, commit, name, checker)
}
func (c *localClient) DiffSymbols(ctx context.Context, repo api.RepoName, commitA, commitB api.CommitID) ([]byte, error) {
	fmt.Println("### localClient.DiffSymbols")
	return c.remoteClient.DiffSymbols(ctx, repo, commitA, commitB)
}
func (c *localClient) ListFiles(ctx context.Context, repo api.RepoName, commit api.CommitID, pattern *regexp.Regexp, checker authz.SubRepoPermissionChecker) ([]string, error) {
	fmt.Println("### localClient.ListFiles")
	return c.remoteClient.ListFiles(ctx, repo, commit, pattern, checker)
}
func (c *localClient) Commits(ctx context.Context, repo api.RepoName, opt CommitsOptions, checker authz.SubRepoPermissionChecker) ([]*gitdomain.Commit, error) {
	fmt.Println("### localClient.Commits")
	return c.remoteClient.Commits(ctx, repo, opt, checker)
}
func (c *localClient) FirstEverCommit(ctx context.Context, repo api.RepoName, checker authz.SubRepoPermissionChecker) (*gitdomain.Commit, error) {
	fmt.Println("### localClient.FirstEverCommit")
	return c.remoteClient.FirstEverCommit(ctx, repo, checker)
}
func (c *localClient) ListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) ([]*gitdomain.Tag, error) {
	fmt.Println("### localClient.ListTags")
	return c.remoteClient.ListTags(ctx, repo, commitObjs...)
}
func (c *localClient) ListDirectoryChildren(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, dirnames []string) (map[string][]string, error) {
	fmt.Println("### localClient.ListDirectoryChildren")
	return c.remoteClient.ListDirectoryChildren(ctx, checker, repo, commit, dirnames)
}
func (c *localClient) Diff(ctx context.Context, opts DiffOptions, checker authz.SubRepoPermissionChecker) (*DiffFileIterator, error) {
	fmt.Println("### localClient.Diff")
	return c.remoteClient.Diff(ctx, opts, checker)
}
func (c *localClient) ReadFile(ctx context.Context, repo api.RepoName, commit api.CommitID, name string, checker authz.SubRepoPermissionChecker) ([]byte, error) {
	fmt.Println("### localClient.ReadFile")
	return c.remoteClient.ReadFile(ctx, repo, commit, name, checker)
}
func (c *localClient) BranchesContaining(ctx context.Context, repo api.RepoName, commit api.CommitID, checker authz.SubRepoPermissionChecker) ([]string, error) {
	fmt.Println("### localClient.BranchesContaining")
	return c.remoteClient.BranchesContaining(ctx, repo, commit, checker)
}
func (c *localClient) RefDescriptions(ctx context.Context, repo api.RepoName, checker authz.SubRepoPermissionChecker, gitObjs ...string) (map[string][]gitdomain.RefDescription, error) {
	fmt.Println("### localClient.RefDescriptions")
	return c.remoteClient.RefDescriptions(ctx, repo, checker, gitObjs...)
}
func (c *localClient) CommitExists(ctx context.Context, repo api.RepoName, id api.CommitID, checker authz.SubRepoPermissionChecker) (bool, error) {
	fmt.Println("### localClient.CommitExists")
	return c.remoteClient.CommitExists(ctx, repo, id, checker)
}
func (c *localClient) CommitsExist(ctx context.Context, repoCommits []api.RepoCommit, checker authz.SubRepoPermissionChecker) ([]bool, error) {
	fmt.Println("### localClient.CommitsExist")
	return c.remoteClient.CommitsExist(ctx, repoCommits, checker)
}
func (c *localClient) Head(ctx context.Context, repo api.RepoName, checker authz.SubRepoPermissionChecker) (string, bool, error) {
	fmt.Println("### localClient.Head")
	return c.remoteClient.Head(ctx, repo, checker)
}
func (c *localClient) CommitDate(ctx context.Context, repo api.RepoName, commit api.CommitID, checker authz.SubRepoPermissionChecker) (string, time.Time, bool, error) {
	fmt.Println("### localClient.CommitDate")
	return c.remoteClient.CommitDate(ctx, repo, commit, checker)
}
func (c *localClient) CommitGraph(ctx context.Context, repo api.RepoName, opts CommitGraphOptions) (_ *gitdomain.CommitGraph, err error) {
	fmt.Println("### localClient.CommitGraph")
	return c.remoteClient.CommitGraph(ctx, repo, opts)
}
func (c *localClient) CommitsUniqueToBranch(ctx context.Context, repo api.RepoName, branchName string, isDefaultBranch bool, maxAge *time.Time, checker authz.SubRepoPermissionChecker) (map[string]time.Time, error) {
	fmt.Println("### localClient.CommitsUniqueToBranch")
	return c.remoteClient.CommitsUniqueToBranch(ctx, repo, branchName, isDefaultBranch, maxAge, checker)
}
func (c *localClient) LsFiles(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, pathspecs ...gitdomain.Pathspec) ([]string, error) {
	if strings.HasPrefix(string(repo), "local/") {
		args := []string{
			"ls-files",
			"-z",
			"--with-tree",
			string(commit),
		}

		if len(pathspecs) > 0 {
			args = append(args, "--")
			for _, pathspec := range pathspecs {
				args = append(args, string(pathspec))
			}
		}

		cmd := exec.Command("git", args...)
		cmd.Dir = filepath.Join(root, strings.TrimPrefix(string(repo), "local/"))

		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		files := strings.Split(string(out), "\x00")
		// Drop trailing empty string
		if len(files) > 0 && files[len(files)-1] == "" {
			files = files[:len(files)-1]
		}
		return filterPaths(ctx, repo, checker, files)
	}
	return c.remoteClient.LsFiles(ctx, checker, repo, commit, pathspecs...)
}
func (c *localClient) GetCommits(ctx context.Context, repoCommits []api.RepoCommit, ignoreErrors bool, checker authz.SubRepoPermissionChecker) ([]*gitdomain.Commit, error) {
	fmt.Println("### localClient.GetCommits")
	return c.remoteClient.GetCommits(ctx, repoCommits, ignoreErrors, checker)
}
func (c *localClient) GetCommit(ctx context.Context, repo api.RepoName, id api.CommitID, opt ResolveRevisionOptions, checker authz.SubRepoPermissionChecker) (*gitdomain.Commit, error) {
	fmt.Println("### localClient.GetCommit")
	return c.remoteClient.GetCommit(ctx, repo, id, opt, checker)
}
func (c *localClient) GetBehindAhead(ctx context.Context, repo api.RepoName, left, right string) (*gitdomain.BehindAhead, error) {
	fmt.Println("### localClient.GetBehindAhead")
	return c.remoteClient.GetBehindAhead(ctx, repo, left, right)
}
func (c *localClient) ContributorCount(ctx context.Context, repo api.RepoName, opt ContributorOptions) ([]*gitdomain.ContributorCount, error) {
	fmt.Println("### localClient.ContributorCount")
	return c.remoteClient.ContributorCount(ctx, repo, opt)
}
func (c *localClient) LogReverseEach(ctx context.Context, repo string, commit string, n int, onLogEntry func(entry gitdomain.LogEntry) error) error {
	fmt.Println("### localClient.LogReverseEach")
	return c.remoteClient.LogReverseEach(ctx, repo, commit, n, onLogEntry)
}
func (c *localClient) RevList(ctx context.Context, repo string, commit string, onCommit func(commit string) (bool, error)) error {
	fmt.Println("### localClient.RevList")
	return c.remoteClient.RevList(ctx, repo, commit, onCommit)
}
func (c *localClient) Addrs() []string {
	return c.remoteClient.Addrs()
}
