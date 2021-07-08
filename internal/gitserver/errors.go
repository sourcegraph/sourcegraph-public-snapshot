package gitserver

import (
	"errors"
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

// IsRevisionNotFound reports if err is a RevisionNotFoundError.
func IsRevisionNotFound(err error) bool {
	// Note we use As instead of Is here to ensure that we do not try
	// compare struct fields for equality; we only care that the type
	// of the error value matches.
	var e *RevisionNotFoundError
	return errors.As(err, &e)
}
