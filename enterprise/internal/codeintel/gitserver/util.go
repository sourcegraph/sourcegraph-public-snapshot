package gitserver

import (
	"bytes"
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

// execGitCommand executes a git command for the given repository by identifier.
func execGitCommand(ctx context.Context, store store.Store, repositoryID int, args ...string) (string, error) {
	repo, err := repositoryIDToRepo(ctx, store, repositoryID)
	if err != nil {
		return "", err
	}

	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = repo
	out, err := cmd.CombinedOutput(ctx)
	return string(bytes.TrimSpace(out)), errors.Wrap(err, "gitserver.Command")
}

// repositoryIDToRepo creates a gitserver.Repo from a repository identifier.
func repositoryIDToRepo(ctx context.Context, store store.Store, repositoryID int) (gitserver.Repo, error) {
	repoName, err := store.RepoName(ctx, repositoryID)
	if err != nil {
		return gitserver.Repo{}, errors.Wrap(err, "store.RepoName")
	}

	return gitserver.Repo{Name: api.RepoName(repoName)}, nil
}
