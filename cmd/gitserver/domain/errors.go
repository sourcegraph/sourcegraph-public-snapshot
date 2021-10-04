package domain

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/api"
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
