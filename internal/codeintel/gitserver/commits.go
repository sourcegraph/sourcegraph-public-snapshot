package gitserver

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

// Head determines the tip commit of the default branch for the given repository.
func Head(db db.DB, repositoryID int) (string, error) {
	// TODO(efritz) - remove dependency on codeintel/db package
	repoName, err := db.RepoName(context.Background(), repositoryID)
	if err != nil {
		return "", err
	}

	cmd := gitserver.DefaultClient.Command("git", "rev-parse", "HEAD")
	cmd.Repo = gitserver.Repo{Name: api.RepoName(repoName)}
	out, err := cmd.CombinedOutput(context.Background())
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(out)), nil
}

// CommitsNear returns a map from a commit to parent commits. The commits populating the
// map are the MaxCommitsPerUpdate closest ancestors from the given commit.
func CommitsNear(db db.DB, repositoryID int, commit string) (map[string][]string, error) {
	// TODO(efritz) - remove dependency on codeintel/db package
	repoName, err := db.RepoName(context.Background(), repositoryID)
	if err != nil {
		return nil, err
	}

	// TODO(efritz) - move this declaration
	const MaxCommitsPerUpdate = 150 // MaxTraversalLimit * 1.5

	cmd := gitserver.DefaultClient.Command("git", "log", "--pretty=%H %P", commit, fmt.Sprintf("-%d", MaxCommitsPerUpdate))
	cmd.Repo = gitserver.Repo{Name: api.RepoName(repoName)}
	out, err := cmd.CombinedOutput(context.Background())
	if err != nil {
		return nil, err
	}

	return parseCommitsNear(strings.Split(string(bytes.TrimSpace(out)), "\n")), nil
}

// parseCommitsNear converts the output of git log into a map from commits to parent commits.
// If a commit is listed but has no ancestors then its parent slice is empty  but is still
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
