package gitserver

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

// Tags returns the git tags associated with the given commit along with a boolean indicating whether
// or not the tag was attached directly to the commit. If no tags exist at or before this commit, the
// tag is an empty string.
func Tags(ctx context.Context, store store.Store, repositoryID int, commit string) (string, bool, error) {
	tag, err := execGitCommand(ctx, store, repositoryID, "tag", "-l", "--points-at", commit)
	if err != nil {
		return "", false, err
	}
	if tag != "" {
		return tag, true, nil
	}

	// git describe --tags will exit with status 128 (fatal: No names found, cannot describe anything)
	// when there are no tags known to the given repo. In order to prevent a gitserver error from
	// occurring, we first check to see if there are any tags and early-exit.
	tags, err := execGitCommand(ctx, store, repositoryID, "tag")
	if err != nil {
		return "", false, err
	}
	if tags == "" {
		return "", false, nil
	}

	tag, err = execGitCommand(ctx, store, repositoryID, "describe", "--tags", "--abbrev=0", commit)
	if err != nil {
		return "", false, err
	}

	return tag, false, nil
}
