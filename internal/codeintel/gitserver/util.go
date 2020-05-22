package gitserver

import (
	"bytes"
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

// execGitCommand executes a git command for the given repository by identifier.
func execGitCommand(ctx context.Context, db db.DB, repositoryID int, args ...string) (string, error) {
	repo, err := repositoryIDToRepo(ctx, db, repositoryID)
	if err != nil {
		return "", err
	}

	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = repo
	out, err := cmd.CombinedOutput(ctx)
	return string(bytes.TrimSpace(out)), errors.Wrap(err, "gitserver.Command")
}

// repositoryIDToRepo creates a gitserver.Repo from a repository identifier.
func repositoryIDToRepo(ctx context.Context, db db.DB, repositoryID int) (gitserver.Repo, error) {
	repoName, err := db.RepoName(ctx, repositoryID)
	if err != nil {
		return gitserver.Repo{}, errors.Wrap(err, "db.RepoName")
	}

	return gitserver.Repo{Name: api.RepoName(repoName)}, nil
}
