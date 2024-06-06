package gitdomain

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// RevisionNotFoundError is an error that reports a revision doesn't exist.
type RevisionNotFoundError struct {
	Repo api.RepoName
	Spec string
}

func (e *RevisionNotFoundError) Error() string {
	return fmt.Sprintf("revision not found: %s@%s", e.Repo, e.Spec)
}

func (e *RevisionNotFoundError) HTTPStatusCode() int {
	return 404
}

func (e *RevisionNotFoundError) NotFound() bool {
	return true
}

// IsRevisionNotFoundError reports if err is a RevisionNotFoundError.
func IsRevisionNotFoundError(err error) bool {
	var e *RevisionNotFoundError
	return errors.As(err, &e)
}

type BadCommitError struct {
	Spec   string
	Commit api.CommitID
	Repo   api.RepoName
}

func (e *BadCommitError) Error() string {
	return fmt.Sprintf("ResolveRevision: got bad commit %q for repo %q at revision %q", e.Commit, e.Repo, e.Spec)
}

// RepoNotExistError is an error that reports a repository doesn't exist.
type RepoNotExistError struct {
	Repo api.RepoName

	// CloneInProgress reports whether the repository is in process of being cloned.
	CloneInProgress bool

	// CloneProgress is a progress message from the running clone command.
	CloneProgress string
}

func (RepoNotExistError) NotFound() bool { return true }

func (e *RepoNotExistError) Error() string {
	if e.CloneInProgress {
		return "repository does not exist (clone in progress): " + string(e.Repo)
	}
	return "repository does not exist: " + string(e.Repo)
}

// IsRepoNotExist reports if err is a RepoNotExistError.
func IsRepoNotExist(err error) bool {
	return errors.HasType[*RepoNotExistError](err)
}

// IsCloneInProgress reports if err is a RepoNotExistError which has a clone
// in progress.
func IsCloneInProgress(err error) bool {
	var e *RepoNotExistError
	return errors.As(err, &e) && e.CloneInProgress
}
