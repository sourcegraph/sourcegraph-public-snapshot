package gitserver

import (
	"cmp"
	"context"
	"io"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type GitserverClient interface {
	// FetchTar returns an io.ReadCloser to a tar archive of a repository at the specified Git
	// remote URL and commit ID. If the error implements "BadRequest() bool", it will be used to
	// determine if the error is a bad request (eg invalid repo).
	FetchTar(context.Context, api.RepoName, api.CommitID, []string) (io.ReadCloser, error)

	// ChangedFiles returns an iterator that yields the paths that have changed between two commits.
	ChangedFiles(context.Context, api.RepoName, string, string) (gitserver.ChangedFilesIterator, error)

	// NewFileReader returns an io.ReadCloser reading from the named file at commit.
	// The caller should always close the reader after use.
	// (If you just need to check a file's existence, use Stat, not a file reader.)
	NewFileReader(ctx context.Context, repoCommitPath types.RepoCommitPath) (io.ReadCloser, error)

	// RevList makes a git rev-list call and iterates through the resulting commits, calling the provided
	// onCommit function for each.
	RevList(ctx context.Context, repo string, commit string, onCommit func(commit string) (shouldContinue bool, err error)) error
}

// Changes are added, deleted, and modified paths.
type Changes struct {
	Added    []string
	Modified []string
	Deleted  []string
}

type gitserverClient struct {
	innerClient gitserver.Client
	operations  *operations
}

func NewClient(observationCtx *observation.Context, inner gitserver.Client) GitserverClient {
	return &gitserverClient{
		innerClient: inner,
		operations:  newOperations(observationCtx),
	}
}

func (c *gitserverClient) FetchTar(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) (_ io.ReadCloser, err error) {
	ctx, _, endObservation := c.operations.fetchTar.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		repo.Attr(),
		commit.Attr(),
		attribute.Int("paths", len(paths)),
	}})
	defer endObservation(1, observation.Args{})

	opts := gitserver.ArchiveOptions{
		Treeish: string(commit),
		Format:  gitserver.ArchiveFormatTar,
		Paths:   paths,
	}

	// Note: the sub-repo perms checker is nil here because we do the sub-repo filtering at a higher level
	return c.innerClient.ArchiveReader(ctx, repo, opts)
}

func (c *gitserverClient) ChangedFiles(ctx context.Context, repo api.RepoName, commitA, commitB string) (iterator gitserver.ChangedFilesIterator, err error) {
	ctx, _, endObservation := c.operations.gitDiff.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		repo.Attr(),
		attribute.String("commitA", commitA),
		attribute.String("commitB", commitB),
	}})
	defer endObservation(1, observation.Args{})

	return c.innerClient.ChangedFiles(ctx, repo, commitA, commitB)
}

func (c *gitserverClient) NewFileReader(ctx context.Context, repoCommitPath types.RepoCommitPath) (io.ReadCloser, error) {
	return c.innerClient.NewFileReader(ctx, api.RepoName(repoCommitPath.Repo), api.CommitID(repoCommitPath.Commit), repoCommitPath.Path)
}

const revListPageSize = 100

func (c *gitserverClient) RevList(ctx context.Context, repo string, commit string, onCommit func(commit string) (shouldContinue bool, err error)) error {
	nextCursor := commit
	for {
		var commits []api.CommitID
		var err error
		commits, nextCursor, err = c.paginatedRevList(ctx, api.RepoName(repo), nextCursor, revListPageSize)
		if err != nil {
			return err
		}
		for _, c := range commits {
			shouldContinue, err := onCommit(string(c))
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}
		}
		if nextCursor == "" {
			return nil
		}
	}
}

func (c *gitserverClient) paginatedRevList(ctx context.Context, repo api.RepoName, commit string, count int) ([]api.CommitID, string, error) {
	commits, err := c.innerClient.Commits(ctx, repo, gitserver.CommitsOptions{
		N:           uint(count + 1),
		Ranges:      []string{cmp.Or(commit, "HEAD")},
		FirstParent: true,
	})
	if err != nil {
		return nil, "", err
	}

	commitIDs := make([]api.CommitID, 0, count+1)

	for _, commit := range commits {
		commitIDs = append(commitIDs, commit.ID)
	}

	var nextCursor string
	if len(commitIDs) > count {
		nextCursor = string(commitIDs[len(commitIDs)-1])
		commitIDs = commitIDs[:count]
	}

	return commitIDs, nextCursor, nil
}
