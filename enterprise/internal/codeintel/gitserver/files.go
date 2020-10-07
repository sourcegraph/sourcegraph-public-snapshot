package gitserver

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// FileExists determines whether a file exists in a particular commit of a repository.
func FileExists(ctx context.Context, store store.Store, repositoryID int, commit, file string) (bool, error) {
	repo, err := repositoryIDToRepo(ctx, store, repositoryID)
	if err != nil {
		return false, err
	}

	if _, err := git.ResolveRevision(ctx, repo, nil, commit, git.ResolveRevisionOptions{}); err != nil {
		return false, errors.Wrap(err, "git.ResolveRevision")
	}

	if _, err := git.Stat(ctx, repo, api.CommitID(commit), file); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// ListFiles returns a list of root-relative file paths matching the given pattern in a particular
// commit of a repository.
func ListFiles(ctx context.Context, store store.Store, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error) {
	out, err := execGitCommand(ctx, store, repositoryID, "ls-tree", "--name-only", "-r", commit, "--")
	if err != nil {
		return nil, err
	}

	var matching []string
	for _, path := range strings.Split(out, "\n") {
		if pattern.MatchString(path) {
			matching = append(matching, path)
		}
	}

	return matching, nil
}
