package gitdomain

import (
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// TODO: We should consistently make all the errors pointer receivers. By being
// pointer receivers it's a compile time error to return them as values and then
// use them as an error type.

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

func (RevisionNotFoundError) NotFound() bool {
	return true
}

type BadCommitError struct {
	Spec   string
	Commit api.CommitID
	Repo   api.RepoName
}

func (e BadCommitError) Error() string {
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
	return errors.HasType(err, &RepoNotExistError{})
}

// IsCloneInProgress reports if err is a RepoNotExistError which has a clone
// in progress.
func IsCloneInProgress(err error) bool {
	var e *RepoNotExistError
	return errors.As(err, &e) && e.CloneInProgress
}
