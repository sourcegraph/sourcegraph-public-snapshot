package gitserver

import (
	"bytes"
	"context"
	"io"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitserverClient interface {
	// FetchTar returns an io.ReadCloser to a tar archive of a repository at the specified Git
	// remote URL and commit ID. If the error implements "BadRequest() bool", it will be used to
	// determine if the error is a bad request (eg invalid repo).
	FetchTar(context.Context, api.RepoName, api.CommitID, []string) (io.ReadCloser, error)

	// GitDiff returns the paths that have changed between two commits.
	GitDiff(context.Context, api.RepoName, api.CommitID, api.CommitID) (Changes, error)

	// GetRepoSize returns the repo size in bytes.
	GetRepoSize(context.Context, api.RepoName) (int64, error)
}

// Changes are added, deleted, and modified paths.
type Changes struct {
	Added    []string
	Modified []string
	Deleted  []string
}

type gitserverClient struct {
	db         database.DB
	operations *operations
}

func NewClient(observationContext *observation.Context) GitserverClient {
	return &gitserverClient{
		operations: newOperations(observationContext),
	}
}

func (c *gitserverClient) FetchTar(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) (_ io.ReadCloser, err error) {
	ctx, _, endObservation := c.operations.fetchTar.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repo", string(repo)),
		log.String("commit", string(commit)),
		log.Int("paths", len(paths)),
	}})
	defer endObservation(1, observation.Args{})

	pathSpecs := []gitserver.Pathspec{}
	for _, path := range paths {
		pathSpecs = append(pathSpecs, gitserver.PathspecLiteral(path))
	}

	opts := gitserver.ArchiveOptions{
		Treeish:   string(commit),
		Format:    "tar",
		Pathspecs: pathSpecs,
	}

	// Note: the sub-repo perms checker is nil here because we do the sub-repo filtering at a higher level
	return git.ArchiveReader(ctx, c.db, nil, repo, opts)
}

func (c *gitserverClient) GitDiff(ctx context.Context, repo api.RepoName, commitA, commitB api.CommitID) (_ Changes, err error) {
	ctx, _, endObservation := c.operations.gitDiff.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repo", string(repo)),
		log.String("commitA", string(commitA)),
		log.String("commitB", string(commitB)),
	}})
	defer endObservation(1, observation.Args{})

	output, err := gitserver.NewClient(c.db).DiffSymbols(ctx, repo, commitA, commitB)

	changes, err := parseGitDiffOutput(output)
	if err != nil {
		return Changes{}, errors.Wrap(err, "failed to parse git diff output")
	}

	return changes, nil
}

func (c *gitserverClient) GetRepoSize(ctx context.Context, repo api.RepoName) (int64, error) {
	repoToInfo, err := gitserver.NewClient(c.db).RepoInfo(ctx, repo)
	if err != nil {
		return 0, err
	}
	info := repoToInfo.Results[repo]
	if info == nil {
		return 0, errors.Errorf("repo %q not found", repo)
	}
	return info.Size, nil
}

var NUL = []byte{0}

// parseGitDiffOutput parses the output of a git diff command, which consists
// of a repeated sequence of `<status> NUL <path> NUL` where NUL is the 0 byte.
func parseGitDiffOutput(output []byte) (changes Changes, _ error) {
	if len(output) == 0 {
		return Changes{}, nil
	}

	slices := bytes.Split(bytes.TrimRight(output, string(NUL)), NUL)
	if len(slices)%2 != 0 {
		return changes, errors.Newf("uneven pairs")
	}

	for i := 0; i < len(slices); i += 2 {
		switch slices[i][0] {
		case 'A':
			changes.Added = append(changes.Added, string(slices[i+1]))
		case 'M':
			changes.Modified = append(changes.Modified, string(slices[i+1]))
		case 'D':
			changes.Deleted = append(changes.Deleted, string(slices[i+1]))
		}
	}

	return changes, nil
}
