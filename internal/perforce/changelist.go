package perforce

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Either git-p4 or p4-fusion could have been used to convert a perforce depot to a git repo. In
// which case the which case the commit message would look like:
//
// [git-p4: depot-paths = "//test-perms/": change = 83725]
// [p4-fusion: depot-paths = "//test-perms/": change = 80972]
//
// NOTE: Do not anchor this pattern to look for the beginning or ending of a line. This ensures that
// we can look for this pattern even when this is not in its own line by itself.
var gitP4Pattern = lazyregexp.New(`\[(?:git-p4|p4-fusion): depot-paths? = "(.*?)"\: change = (\d+)\]`)

func GetP4ChangelistID(body string) (string, error) {
	matches := gitP4Pattern.FindStringSubmatch(body)
	if len(matches) != 3 {
		return "", errors.Newf("failed to retrieve changelist ID from commit body: %q", body)
	}

	return matches[2], nil
}

// ChangelistNotFoundError is an error that reports a revision doesn't exist.
type ChangelistNotFoundError struct {
	Repo api.RepoName
	ID   string
}

func (e *ChangelistNotFoundError) Error() string {
	return fmt.Sprintf("revision not found: %s@%s", e.Repo, e.ID)
}
