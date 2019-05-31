package git

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
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

// IsRevisionNotFound reports if err or its cause is a RevisionNotFoundError.
func IsRevisionNotFound(err error) bool {
	_, ok := err.(*RevisionNotFoundError)
	if !ok {
		err = errors.Cause(err)
		_, ok := err.(*RevisionNotFoundError)
		return ok
	}
	return ok
}
