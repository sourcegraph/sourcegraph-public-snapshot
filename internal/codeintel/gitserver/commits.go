package gitserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

// TODO(efritz) - move this declaration (MaxCommitsPerUpdate = MaxTraversalLimit * 1.5)
const MaxCommitsPerUpdate = 150

// Head determines the tip commit of the default branch for the given repository.
func Head(ctx context.Context, db db.DB, repositoryID int) (string, error) {
	return execGitCommand(ctx, db, repositoryID, "rev-parse", "HEAD")
}

// CommitsNear returns a map from a commit to parent commits. The commits populating the
// map are the MaxCommitsPerUpdate closest ancestors from the given commit.
func CommitsNear(ctx context.Context, db db.DB, repositoryID int, commit string) (map[string][]string, error) {
	out, err := execGitCommand(ctx, db, repositoryID, "log", "--pretty=%H %P", commit, fmt.Sprintf("-%d", MaxCommitsPerUpdate))
	if err != nil {
		return nil, err
	}

	return parseCommitsNear(strings.Split(out, "\n")), nil
}

// parseCommitsNear converts the output of git log into a map from commits to parent commits.
// If a commit is listed but has no ancestors then its parent slice is empty but is still
// present in the map.
func parseCommitsNear(pair []string) map[string][]string {
	commits := map[string][]string{}

	for _, pair := range pair {
		line := strings.TrimSpace(pair)
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		commits[parts[0]] = parts[1:]
	}

	return commits
}
