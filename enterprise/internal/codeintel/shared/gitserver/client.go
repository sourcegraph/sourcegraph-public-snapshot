package gitserver

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/grafana/regexp"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Client struct {
	gitserverClient gitserver.Client
	dbStore         *store
	operations      *operations
}

func New(observationCtx *observation.Context, db database.DB) *Client {
	observationCtx = observation.ContextWithLogger(log.NoOp(), observationCtx)
	operations := newOperations(observationCtx)

	return &Client{
		gitserverClient: gitserver.NewClient(db),
		dbStore:         newWithDB(db),
		operations:      operations,
	}
}

func NewWithGitserverClient(observationCtx *observation.Context, db database.DB, gitserverClient gitserver.Client) *Client {
	return &Client{
		gitserverClient: gitserverClient,
		dbStore:         newWithDB(db),
		operations:      newOperations(observationCtx),
	}
}

func (c *Client) ArchiveReader(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, options gitserver.ArchiveOptions) (io.ReadCloser, error) {
	return c.gitserverClient.ArchiveReader(ctx, checker, repo, options)
}

func (c *Client) RequestRepoUpdate(ctx context.Context, name api.RepoName, t time.Duration) (*protocol.RepoUpdateResponse, error) {
	return c.gitserverClient.RequestRepoUpdate(ctx, name, t)
}

func (c *Client) DiffPath(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, sourceCommit, targetCommit, path string) ([]*diff.Hunk, error) {
	return c.gitserverClient.DiffPath(ctx, checker, repo, sourceCommit, targetCommit, path)
}

// CommitExists determines if the given commit exists in the given repository.
func (c *Client) CommitExists(ctx context.Context, repositoryID int, commit string) (_ bool, err error) {
	ctx, _, endObservation := c.operations.commitExists.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("repositoryID", repositoryID),
		otlog.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return false, err
	}
	return c.gitserverClient.CommitExists(ctx, repo, api.CommitID(commit), authz.DefaultSubRepoPermsChecker)
}

type RepositoryCommit struct {
	RepositoryID int
	Commit       string
}

// CommitsExist determines if the given commits exists in the given repositories. This method returns a
// slice of the same size as the input slice, true indicating that the commit at the symmetric index exists.
func (c *Client) CommitsExist(ctx context.Context, commits []RepositoryCommit) (_ []bool, err error) {
	ctx, _, endObservation := c.operations.commitsExist.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("numCommits", len(commits)),
	}})
	defer endObservation(1, observation.Args{})

	repositoryIDMap := map[int]struct{}{}
	for _, rc := range commits {
		repositoryIDMap[rc.RepositoryID] = struct{}{}
	}
	repositoryIDs := make([]int, 0, len(repositoryIDMap))
	for repositoryID := range repositoryIDMap {
		repositoryIDs = append(repositoryIDs, repositoryID)
	}

	repositoryNames, err := c.repositoryIDsToRepos(ctx, repositoryIDs...)
	if err != nil {
		return nil, err
	}

	// Build the batch request to send to gitserver. Because we only add repo/commit
	// pairs that are resolvable to a repo name, we may skip some of the inputs here.
	// We track the indexes we're sending to gitserver so we can spread the response
	// back to the correct indexes the caller is expecting. Anything not resolvable
	// to a repository name will implicity have a false value in the returned slice.

	repoCommits := make([]api.RepoCommit, 0, len(commits))
	originalIndexes := make([]int, 0, len(commits))

	for i, rc := range commits {
		repoName, ok := repositoryNames[rc.RepositoryID]
		if !ok {
			continue
		}

		repoCommits = append(repoCommits, api.RepoCommit{
			Repo:     repoName,
			CommitID: api.CommitID(rc.Commit),
		})

		originalIndexes = append(originalIndexes, i)
	}

	exists, err := c.gitserverClient.CommitsExist(ctx, repoCommits, authz.DefaultSubRepoPermsChecker)
	if err != nil {
		return nil, err
	}
	if len(exists) != len(repoCommits) {
		// Add assertion here so that the blast radius of new or newly discovered errors southbound
		// from the internal/vcs/git package does not leak into code intelligence. The existing callers
		// of this method panic when this assertion is not met. Describing the error in more detail here
		// will not cause destruction outside of the particular user-request in which this assertion
		// was not true.
		return nil, errors.Newf("expected slice returned from git.CommitsExist to have len %d, but has len %d", len(repoCommits), len(exists))
	}

	// Spread the response back to the correct indexes the caller is expecting. Each value in the
	// response from gitserver belongs to some index in the original commits slice. We re-map these
	// values and leave all other values implicitly false (these repo name were not resolvable).
	out := make([]bool, len(commits))
	for i, e := range exists {
		out[originalIndexes[i]] = e
	}

	return out, nil
}

// Head determines the tip commit of the default branch for the given repository. If no HEAD revision exists
// for the given repository (which occurs with empty repositories), a false-valued flag is returned along with
// a nil error and empty revision.
func (c *Client) Head(ctx context.Context, repositoryID int) (_ string, revisionExists bool, err error) {
	ctx, _, endObservation := c.operations.head.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return "", false, err
	}

	return c.gitserverClient.Head(ctx, repo, authz.DefaultSubRepoPermsChecker)
}

func (c *Client) HeadFromName(ctx context.Context, repo api.RepoName) (_ string, revisionExists bool, err error) {
	return c.gitserverClient.Head(ctx, repo, authz.DefaultSubRepoPermsChecker)
}

// CommitDate returns the time that the given commit was committed. If the given revision does not exist,
// a false-valued flag is returned along with a nil error and zero-valued time.
func (c *Client) CommitDate(ctx context.Context, repositoryID int, commit string) (_ string, _ time.Time, revisionExists bool, err error) {
	ctx, _, endObservation := c.operations.commitDate.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("repositoryID", repositoryID),
		otlog.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return "", time.Time{}, false, nil
	}

	rev, tm, ok, err := c.gitserverClient.CommitDate(ctx, repo, api.CommitID(commit), authz.DefaultSubRepoPermsChecker)
	if err == nil {
		return rev, tm, ok, nil
	}

	// If the repo doesn't exist don't bother trying to resolve the commit.
	// Otherwise, if we're returning an error, try to resolve revision that was the
	// target of the command. If the revision fails to resolve, we return an instance
	// of a RevisionNotFoundError error instead of an "exit 128".
	if !gitdomain.IsRepoNotExist(err) {
		if _, err := c.gitserverClient.ResolveRevision(ctx, repo, commit, gitserver.ResolveRevisionOptions{}); err != nil {
			return "", time.Time{}, false, errors.Wrap(err, "git.ResolveRevision")
		}
	}

	// If we didn't expect a particular revision to exist, or we did but it
	// resolved without error, return the original error as the command had
	// failed for another reason.
	return "", time.Time{}, false, errors.Wrap(err, "git.CommitDate")
}

// CommitGraph returns the commit graph for the given repository as a mapping from a commit
// to its parents. If a commit is supplied, the returned graph will be rooted at the given
// commit. If a non-zero limit is supplied, at most that many commits will be returned.
func (c *Client) CommitGraph(ctx context.Context, repositoryID int, opts gitserver.CommitGraphOptions) (_ *gitdomain.CommitGraph, err error) {
	ctx, _, endObservation := c.operations.commitGraph.With(ctx, &err, observation.Args{
		LogFields: append([]otlog.Field{otlog.Int("repositoryID", repositoryID)}, opts.LogFields()...),
	})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	gitserverClient := c.gitserverClient
	g, err := gitserverClient.CommitGraph(ctx, repo, opts)
	if err == nil {
		return g, nil
	}

	// If the repo doesn't exist don't bother trying to resolve the commit.
	// Otherwise, if we're returning an error, try to resolve revision that was the
	// target of the command. If the revision fails to resolve, we return an instance
	// of a RevisionNotFoundError error instead of an "exit 128".
	if !gitdomain.IsRepoNotExist(err) && opts.Commit != "" {
		if _, err := gitserverClient.ResolveRevision(ctx, repo, opts.Commit, gitserver.ResolveRevisionOptions{}); err != nil {
			return nil, errors.Wrap(err, "git.ResolveRevision")
		}
	}

	// If we didn't expect a particular revision to exist, or we did but it
	// resolved without error, return the original error as the command had
	// failed for another reason.
	return nil, errors.Wrap(err, "git.CommitGraph")
}

// RefDescriptions returns a map from commits to descriptions of the tip of each
// branch and tag of the given repository. If any git objects are provided, it will
// only populate entries for descriptions pointing at the given git objects.
func (c *Client) RefDescriptions(ctx context.Context, repositoryID int, pointedAt ...string) (_ map[string][]gitdomain.RefDescription, err error) {
	ctx, _, endObservation := c.operations.refDescriptions.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	return c.gitserverClient.RefDescriptions(ctx, repo, authz.DefaultSubRepoPermsChecker, pointedAt...)
}

// CommitsUniqueToBranch returns a map from commits that exist on a particular branch in the given repository to
// their committer date. This set of commits is determined by listing `{branchName} ^HEAD`, which is interpreted
// as: all commits on {branchName} not also on the tip of the default branch. If the supplied branch name is the
// default branch, then this method instead returns all commits reachable from HEAD.
func (c *Client) CommitsUniqueToBranch(ctx context.Context, repositoryID int, branchName string, isDefaultBranch bool, maxAge *time.Time) (_ map[string]time.Time, err error) {
	ctx, _, endObservation := c.operations.commitsUniqueToBranch.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("repositoryID", repositoryID),
		otlog.String("branchName", branchName),
		otlog.Bool("isDefaultBranch", isDefaultBranch),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	return c.gitserverClient.CommitsUniqueToBranch(ctx, repo, branchName, isDefaultBranch, maxAge, authz.DefaultSubRepoPermsChecker)
}

// branchesContaining returns a map from branch names to branch tip hashes for each branch
// containing the given commit.
func (c *Client) branchesContaining(ctx context.Context, repositoryID int, commit string) ([]string, error) {
	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	return c.gitserverClient.BranchesContaining(ctx, repo, api.CommitID(commit), authz.DefaultSubRepoPermsChecker)
}

// DefaultBranchContains tells if the default branch contains the given commit ID.
func (c *Client) DefaultBranchContains(ctx context.Context, repositoryID int, commit string) (bool, error) {
	// Determine default branch name.
	descriptions, err := c.RefDescriptions(ctx, repositoryID)
	if err != nil {
		return false, errors.Wrap(err, "RefDescriptions")
	}
	var defaultBranchName string
	for _, descriptions := range descriptions {
		for _, ref := range descriptions {
			if ref.IsDefaultBranch {
				defaultBranchName = ref.Name
				break
			}
		}
	}

	// Determine if branch contains commit.
	branches, err := c.branchesContaining(ctx, repositoryID, commit)
	if err != nil {
		return false, errors.Wrap(err, "BranchesContaining")
	}
	for _, branch := range branches {
		if branch == defaultBranchName {
			return true, nil
		}
	}
	return false, nil
}

// RawContents returns the contents of a file in a particular commit of a repository.
func (c *Client) RawContents(ctx context.Context, repositoryID int, commit, file string) (_ []byte, err error) {
	ctx, _, endObservation := c.operations.rawContents.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("repositoryID", repositoryID),
		otlog.String("commit", commit),
		otlog.String("file", file),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	out, err := c.gitserverClient.ReadFile(ctx, repo, api.CommitID(commit), file, authz.DefaultSubRepoPermsChecker)
	if err == nil {
		return out, nil
	}

	// If the repo doesn't exist don't bother trying to resolve the commit.
	// Otherwise, if we're returning an error, try to resolve revision that was the
	// target of the command. If the revision fails to resolve, we return an instance
	// of a RevisionNotFoundError error instead of an "exit 128".
	if !gitdomain.IsRepoNotExist(err) {
		if _, err := c.gitserverClient.ResolveRevision(ctx, repo, commit, gitserver.ResolveRevisionOptions{}); err != nil {
			return nil, errors.Wrap(err, "git.ResolveRevision")
		}
	}

	// If we didn't expect a particular revision to exist, or we did but it
	// resolved without error, return the original error as the command had
	// failed for another reason.
	return nil, errors.Wrap(err, "git.ReadFile")
}

// DirectoryChildren determines all children known to git for the given directory names via an invocation
// of git ls-tree. The keys of the resulting map are the input (unsanitized) dirnames, and the value of
// that key are the files nested under that directory.
func (c *Client) DirectoryChildren(ctx context.Context, repositoryID int, commit string, dirnames []string) (_ map[string][]string, err error) {
	ctx, _, endObservation := c.operations.directoryChildren.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("repositoryID", repositoryID),
		otlog.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	children, err := c.gitserverClient.ListDirectoryChildren(ctx, authz.DefaultSubRepoPermsChecker, repo, api.CommitID(commit), dirnames)
	if err == nil {
		return children, err
	}

	// If the repo doesn't exist don't bother trying to resolve the commit.
	// Otherwise, if we're returning an error, try to resolve revision that was the
	// target of the command. If the revision fails to resolve, we return an instance
	// of a RevisionNotFoundError error instead of an "exit 128".
	if !gitdomain.IsRepoNotExist(err) {
		if _, err := c.gitserverClient.ResolveRevision(ctx, repo, commit, gitserver.ResolveRevisionOptions{}); err != nil {
			return nil, errors.Wrap(err, "git.ResolveRevision")
		}
	}

	// If we didn't expect a particular revision to exist, or we did but it
	// resolved without error, return the original error as the command had
	// failed for another reason.
	return nil, errors.Wrap(err, "git.ListDirectoryChildren")
}

// FileExists determines whether a file exists in a particular commit of a repository.
func (c *Client) FileExists(ctx context.Context, repositoryID int, commit, file string) (_ bool, err error) {
	ctx, _, endObservation := c.operations.fileExists.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("repositoryID", repositoryID),
		otlog.String("commit", commit),
		otlog.String("file", file),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return false, err
	}

	if _, err := c.gitserverClient.ResolveRevision(ctx, repo, commit, gitserver.ResolveRevisionOptions{}); err != nil {
		return false, errors.Wrap(err, "git.ResolveRevision")
	}

	if _, err := c.gitserverClient.Stat(ctx, authz.DefaultSubRepoPermsChecker, repo, api.CommitID(commit), file); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// ListFiles returns a list of root-relative file paths matching the given pattern in a particular
// commit of a repository.
func (c *Client) ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) (_ []string, err error) {
	ctx, _, endObservation := c.operations.listFiles.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("repositoryID", repositoryID),
		otlog.String("commit", commit),
		otlog.String("pattern", pattern.String()),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	return c.ListFilesForRepo(ctx, repo, commit, pattern)
}

func (c *Client) ListFilesForRepo(ctx context.Context, repo api.RepoName, commit string, pattern *regexp.Regexp) (_ []string, err error) {
	matching, err := c.gitserverClient.ListFiles(ctx, repo, api.CommitID(commit), pattern, authz.DefaultSubRepoPermsChecker)
	if err == nil {
		return matching, nil
	}

	// If the repo doesn't exist don't bother trying to resolve the commit.
	// Otherwise, if we're returning an error, try to resolve revision that was the
	// target of the command. If the revision fails to resolve, we return an instance
	// of a RevisionNotFoundError error instead of an "exit 128".
	if !gitdomain.IsRepoNotExist(err) {
		if _, err := c.gitserverClient.ResolveRevision(ctx, repo, commit, gitserver.ResolveRevisionOptions{}); err != nil {
			return nil, errors.Wrap(err, "git.ResolveRevision")
		}
	}

	// If we didn't expect a particular revision to exist, or we did but it
	// resolved without error, return the original error as the command had
	// failed for another reason.
	return nil, errors.Wrap(err, "git.ListFiles")
}

// ResolveRevision returns the absolute commit for a commit-ish spec.
func (c *Client) ResolveRevision(ctx context.Context, repositoryID int, versionString string) (commitID api.CommitID, err error) {
	ctx, _, endObservation := c.operations.resolveRevision.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("repositoryID", repositoryID),
		otlog.String("versionString", versionString),
	}})
	defer endObservation(1, observation.Args{})

	repoName, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return "", err
	}

	commitID, err = c.gitserverClient.ResolveRevision(ctx, repoName, versionString, gitserver.ResolveRevisionOptions{})
	if err != nil {
		return "", errors.Wrap(err, "git.ResolveRevision")
	}

	return commitID, nil
}

func (c *Client) ListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error) {
	ctx, _, endObservation := c.operations.listTags.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("commitObjs", strings.Join(commitObjs, ",")),
	}})
	defer endObservation(1, observation.Args{})

	tags, err := c.gitserverClient.ListTags(ctx, repo, commitObjs...)
	if err != nil {
		return nil, errors.Wrap(err, "git.ListTags")
	}
	return tags, nil
}

// repositoryIDToRepo creates a api.RepoName from a repository identifier.
func (c *Client) repositoryIDToRepo(ctx context.Context, repositoryID int) (api.RepoName, error) {
	repoName, err := c.dbStore.RepoName(ctx, repositoryID)
	if err != nil {
		return "", errors.Wrap(err, "dbstore.RepoName")
	}

	return api.RepoName(repoName), nil
}

// repositoryIDsToRepos creates a map from repository identifiers to api.RepoNames.
func (c *Client) repositoryIDsToRepos(ctx context.Context, repositoryIDs ...int) (map[int]api.RepoName, error) {
	names, err := c.dbStore.RepoNames(ctx, repositoryIDs...)
	if err != nil {
		return nil, err
	}

	repoNames := make(map[int]api.RepoName, len(names))
	for repositoryID, name := range names {
		repoNames[repositoryID] = api.RepoName(name)
	}

	return repoNames, nil
}
