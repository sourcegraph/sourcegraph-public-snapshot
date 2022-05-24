package gitserver

import (
	"context"
	"os"
	"time"

	"github.com/grafana/regexp"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Client struct {
	db         database.DB
	dbStore    DBStore
	operations *operations
}

func New(db database.DB, dbStore DBStore, observationContext *observation.Context) *Client {
	return &Client{
		db:         db,
		dbStore:    dbStore,
		operations: newOperations(observationContext),
	}
}

// CommitExists determines if the given commit exists in the given repository.
func (c *Client) CommitExists(ctx context.Context, repositoryID int, commit string) (_ bool, err error) {
	ctx, _, endObservation := c.operations.commitExists.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return false, err
	}
	return git.CommitExists(ctx, c.db, repo, api.CommitID(commit), authz.DefaultSubRepoPermsChecker)
}

type RepositoryCommit struct {
	RepositoryID int
	Commit       string
}

// CommitsExist determines if the given commits exists in the given repositories. This method returns a
// slice of the same size as the input slice, true indicating that the commit at the symmetric index exists.
func (c *Client) CommitsExist(ctx context.Context, commits []RepositoryCommit) (_ []bool, err error) {
	ctx, _, endObservation := c.operations.commitsExist.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numCommits", len(commits)),
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

	exists, err := git.CommitsExist(ctx, c.db, repoCommits, authz.DefaultSubRepoPermsChecker)
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
	ctx, _, endObservation := c.operations.head.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return "", false, err
	}

	return git.Head(ctx, c.db, repo, authz.DefaultSubRepoPermsChecker)
}

// CommitDate returns the time that the given commit was committed. If the given revision does not exist,
// a false-valued flag is returned along with a nil error and zero-valued time.
func (c *Client) CommitDate(ctx context.Context, repositoryID int, commit string) (_ string, _ time.Time, revisionExists bool, err error) {
	ctx, _, endObservation := c.operations.commitDate.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return "", time.Time{}, false, nil
	}

	db := c.db
	rev, tm, ok, err := git.CommitDate(ctx, db, repo, api.CommitID(commit), authz.DefaultSubRepoPermsChecker)
	if err == nil {
		return rev, tm, ok, nil
	}

	// If the repo doesn't exist don't bother trying to resolve the commit.
	// Otherwise, if we're returning an error, try to resolve revision that was the
	// target of the command. If the revision fails to resolve, we return an instance
	// of a RevisionNotFoundError error instead of an "exit 128".
	if !gitdomain.IsRepoNotExist(err) {
		if _, err := gitserver.NewClient(c.db).ResolveRevision(ctx, repo, commit, gitserver.ResolveRevisionOptions{}); err != nil {
			return "", time.Time{}, false, errors.Wrap(err, "git.ResolveRevision")
		}
	}

	// If we didn't expect a particular revision to exist, or we did but it
	// resolved without error, return the original error as the command had
	// failed for another reason.
	return "", time.Time{}, false, errors.Wrap(err, "git.CommitDate")
}

func (c *Client) RepoInfo(ctx context.Context, repos ...api.RepoName) (_ map[api.RepoName]*protocol.RepoInfo, err error) {
	ctx, _, endObservation := c.operations.repoInfo.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numRepos", len(repos)),
	}})
	defer endObservation(1, observation.Args{})

	resp, err := gitserver.NewClient(c.db).RepoInfo(ctx, repos...)
	if resp == nil {
		return nil, err
	}

	return resp.Results, err
}

// CommitGraph returns the commit graph for the given repository as a mapping from a commit
// to its parents. If a commit is supplied, the returned graph will be rooted at the given
// commit. If a non-zero limit is supplied, at most that many commits will be returned.
func (c *Client) CommitGraph(ctx context.Context, repositoryID int, opts gitserver.CommitGraphOptions) (_ *gitdomain.CommitGraph, err error) {
	ctx, _, endObservation := c.operations.commitGraph.With(ctx, &err, observation.Args{
		LogFields: append([]log.Field{log.Int("repositoryID", repositoryID)}, opts.LogFields()...),
	})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	gitserverClient := gitserver.NewClient(c.db)
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
	ctx, _, endObservation := c.operations.refDescriptions.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	return git.RefDescriptions(ctx, c.db, repo, authz.DefaultSubRepoPermsChecker, pointedAt...)
}

// CommitsUniqueToBranch returns a map from commits that exist on a particular branch in the given repository to
// their committer date. This set of commits is determined by listing `{branchName} ^HEAD`, which is interpreted
// as: all commits on {branchName} not also on the tip of the default branch. If the supplied branch name is the
// default branch, then this method instead returns all commits reachable from HEAD.
func (c *Client) CommitsUniqueToBranch(ctx context.Context, repositoryID int, branchName string, isDefaultBranch bool, maxAge *time.Time) (_ map[string]time.Time, err error) {
	ctx, _, endObservation := c.operations.commitsUniqueToBranch.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("branchName", branchName),
		log.Bool("isDefaultBranch", isDefaultBranch),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	return git.CommitsUniqueToBranch(ctx, c.db, repo, branchName, isDefaultBranch, maxAge, authz.DefaultSubRepoPermsChecker)
}

// BranchesContaining returns a map from branch names to branch tip hashes for each branch
// containing the given commit.
func (c *Client) BranchesContaining(ctx context.Context, db database.DB, repositoryID int, commit string) ([]string, error) {
	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	return git.BranchesContaining(ctx, db, repo, api.CommitID(commit), authz.DefaultSubRepoPermsChecker)
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
	branches, err := c.BranchesContaining(ctx, c.db, repositoryID, commit)
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
	ctx, _, endObservation := c.operations.rawContents.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
		log.String("file", file),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	db := c.db
	out, err := git.ReadFile(ctx, db, repo, api.CommitID(commit), file, authz.DefaultSubRepoPermsChecker)
	if err == nil {
		return out, nil
	}

	// If the repo doesn't exist don't bother trying to resolve the commit.
	// Otherwise, if we're returning an error, try to resolve revision that was the
	// target of the command. If the revision fails to resolve, we return an instance
	// of a RevisionNotFoundError error instead of an "exit 128".
	if !gitdomain.IsRepoNotExist(err) {
		if _, err := gitserver.NewClient(c.db).ResolveRevision(ctx, repo, commit, gitserver.ResolveRevisionOptions{}); err != nil {
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
	ctx, _, endObservation := c.operations.directoryChildren.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	children, err := gitserver.NewClient(c.db).ListDirectoryChildren(ctx, authz.DefaultSubRepoPermsChecker, repo, api.CommitID(commit), dirnames)
	if err == nil {
		return children, err
	}

	// If the repo doesn't exist don't bother trying to resolve the commit.
	// Otherwise, if we're returning an error, try to resolve revision that was the
	// target of the command. If the revision fails to resolve, we return an instance
	// of a RevisionNotFoundError error instead of an "exit 128".
	if !gitdomain.IsRepoNotExist(err) {
		if _, err := gitserver.NewClient(c.db).ResolveRevision(ctx, repo, commit, gitserver.ResolveRevisionOptions{}); err != nil {
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
	ctx, _, endObservation := c.operations.fileExists.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
		log.String("file", file),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return false, err
	}

	db := c.db
	if _, err := gitserver.NewClient(c.db).ResolveRevision(ctx, repo, commit, gitserver.ResolveRevisionOptions{}); err != nil {
		return false, errors.Wrap(err, "git.ResolveRevision")
	}

	if _, err := git.Stat(ctx, db, authz.DefaultSubRepoPermsChecker, repo, api.CommitID(commit), file); err != nil {
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
	ctx, _, endObservation := c.operations.listFiles.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
		log.String("pattern", pattern.String()),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	matching, err := gitserver.NewClient(c.db).ListFiles(ctx, repo, api.CommitID(commit), pattern, authz.DefaultSubRepoPermsChecker)
	if err == nil {
		return matching, nil
	}

	// If the repo doesn't exist don't bother trying to resolve the commit.
	// Otherwise, if we're returning an error, try to resolve revision that was the
	// target of the command. If the revision fails to resolve, we return an instance
	// of a RevisionNotFoundError error instead of an "exit 128".
	if !gitdomain.IsRepoNotExist(err) {
		if _, err := gitserver.NewClient(c.db).ResolveRevision(ctx, repo, commit, gitserver.ResolveRevisionOptions{}); err != nil {
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
	ctx, _, endObservation := c.operations.resolveRevision.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("versionString", versionString),
	}})
	defer endObservation(1, observation.Args{})

	repoName, err := c.repositoryIDToRepo(ctx, repositoryID)
	if err != nil {
		return "", err
	}

	commitID, err = gitserver.NewClient(c.db).ResolveRevision(ctx, repoName, versionString, gitserver.ResolveRevisionOptions{})
	if err != nil {
		return "", errors.Wrap(err, "git.ResolveRevision")
	}

	return commitID, nil
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
