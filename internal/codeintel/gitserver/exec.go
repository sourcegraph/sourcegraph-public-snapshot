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
	// TODO(efritz) - remove dependency on codeintel/db package
	repoName, err := db.RepoName(ctx, repositoryID)
	if err != nil {
		return "", errors.Wrap(err, "db.RepoName")
	}

	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = gitserver.Repo{Name: api.RepoName(repoName)}
	out, err := cmd.CombinedOutput(ctx)
	return string(bytes.TrimSpace(out)), errors.Wrap(err, "gitserver.Command")
}
