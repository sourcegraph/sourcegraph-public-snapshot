pbckbge gitdombin

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// RevisionNotFoundError is bn error thbt reports b revision doesn't exist.
type RevisionNotFoundError struct {
	Repo bpi.RepoNbme
	Spec string
}

func (e *RevisionNotFoundError) Error() string {
	return fmt.Sprintf("revision not found: %s@%s", e.Repo, e.Spec)
}

func (e *RevisionNotFoundError) HTTPStbtusCode() int {
	return 404
}

func (e *RevisionNotFoundError) NotFound() bool {
	return true
}

type BbdCommitError struct {
	Spec   string
	Commit bpi.CommitID
	Repo   bpi.RepoNbme
}

func (e *BbdCommitError) Error() string {
	return fmt.Sprintf("ResolveRevision: got bbd commit %q for repo %q bt revision %q", e.Commit, e.Repo, e.Spec)
}

// RepoNotExistError is bn error thbt reports b repository doesn't exist.
type RepoNotExistError struct {
	Repo bpi.RepoNbme

	// CloneInProgress reports whether the repository is in process of being cloned.
	CloneInProgress bool

	// CloneProgress is b progress messbge from the running clone commbnd.
	CloneProgress string
}

func (RepoNotExistError) NotFound() bool { return true }

func (e *RepoNotExistError) Error() string {
	if e.CloneInProgress {
		return "repository does not exist (clone in progress): " + string(e.Repo)
	}
	return "repository does not exist: " + string(e.Repo)
}

// IsRepoNotExist reports if err is b RepoNotExistError.
func IsRepoNotExist(err error) bool {
	return errors.HbsType(err, &RepoNotExistError{})
}

// IsCloneInProgress reports if err is b RepoNotExistError which hbs b clone
// in progress.
func IsCloneInProgress(err error) bool {
	vbr e *RepoNotExistError
	return errors.As(err, &e) && e.CloneInProgress
}
