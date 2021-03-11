// Package errcode maps Go errors to HTTP status codes as well as other useful
// functions for inspecting errors.
package errcode

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/schema"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

// HTTP returns the most appropriate HTTP status code that describes
// err. It contains a hard-coded list of error types and error values
// (such as mapping store.RepoNotFoundError to NotFound) and
// heuristics (such as mapping os.IsNotExist-satisfying errors to
// NotFound). All other errors are mapped to HTTP 500 Internal Server
// Error.
func HTTP(err error) int {
	if err == nil {
		return http.StatusOK
	}

	switch err {
	case context.DeadlineExceeded:
		return http.StatusRequestTimeout
	}

	if vcs.IsCloneInProgress(err) || strings.Contains(err.Error(), (&vcs.RepoNotExistError{CloneInProgress: true}).Error()) {
		return http.StatusAccepted
	} else if vcs.IsRepoNotExist(err) || strings.Contains(err.Error(), (&vcs.RepoNotExistError{}).Error()) {
		return http.StatusNotFound
	}

	switch e := err.(type) {
	case interface {
		HTTPStatusCode() int
	}:
		return e.HTTPStatusCode()
	case schema.ConversionError:
		return http.StatusBadRequest
	case schema.MultiError:
		return http.StatusBadRequest
	}

	if os.IsNotExist(err) {
		return http.StatusNotFound
	} else if os.IsPermission(err) {
		return http.StatusForbidden
	} else if IsNotFound(err) {
		return http.StatusNotFound
	} else if IsBadRequest(err) {
		return http.StatusBadRequest
	}

	return http.StatusInternalServerError
}

type HTTPErr struct {
	Status int   // HTTP status code.
	Err    error // Optional reason for the HTTP error.
}

func (err *HTTPErr) Error() string {
	if err.Err != nil {
		return fmt.Sprintf("status %d, reason %s", err.Status, err.Err)
	}
	return fmt.Sprintf("Status %d", err.Status)
}

func (err *HTTPErr) HTTPStatusCode() int { return err.Status }

func IsHTTPErrorCode(err error, statusCode int) bool {
	return HTTP(err) == statusCode
}

// Mock is a convenience error which makes it easy to satisfy the optional
// interfaces errors implement.
type Mock struct {
	// Message is the return value for Error() string
	Message string

	// IsNotFound is the return value for NotFound
	IsNotFound bool
}

func (e *Mock) Error() string {
	return e.Message
}

func (e *Mock) NotFound() bool {
	return e.IsNotFound
}

// IsNotFound will check if err or one of its causes is a not found
// error. Note: This will not check os.IsNotExist, but rather is returned by
// methods like Repo.Get when the repo is not found. It will also *not* map
// HTTPStatusCode into not found.
func IsNotFound(err error) bool {
	type notFounder interface {
		NotFound() bool
	}
	return isErrorPredicate(err, func(err error) bool {
		e, ok := err.(notFounder)
		return ok && e.NotFound()
	})
}

// IsUnauthorized will check if err or one of its causes is an unauthorized
// error.
func IsUnauthorized(err error) bool {
	type unauthorizeder interface {
		Unauthorized() bool
	}
	return isErrorPredicate(err, func(err error) bool {
		e, ok := err.(unauthorizeder)
		return ok && e.Unauthorized()
	})
}

// IsAccountSuspended will check if err or one of its causes was due to the
// account being suspended
func IsAccountSuspended(err error) bool {
	type accountSuspendeder interface {
		AccountSuspended() bool
	}
	return isErrorPredicate(err, func(err error) bool {
		e, ok := err.(accountSuspendeder)
		return ok && e.AccountSuspended()
	})
}

// IsBadRequest will check if err or one of its causes is a bad request.
func IsBadRequest(err error) bool {
	type badRequester interface {
		BadRequest() bool
	}
	return isErrorPredicate(err, func(err error) bool {
		badrequest, ok := err.(badRequester)
		return ok && badrequest.BadRequest()
	})
}

// IsTemporary will check if err or one of its causes is temporary. A
// temporary error can be retried. Many errors in the go stdlib implement the
// temporary interface.
func IsTemporary(err error) bool {
	type temporaryer interface {
		Temporary() bool
	}
	return isErrorPredicate(err, func(err error) bool {
		e, ok := err.(temporaryer)
		return ok && e.Temporary()
	})
}

// IsTimeout will check if err or one of its causes is a timeout. Many errors
// in the go stdlib implement the timeout interface.
func IsTimeout(err error) bool {
	type timeouter interface {
		Timeout() bool
	}
	return isErrorPredicate(err, func(err error) bool {
		e, ok := err.(timeouter)
		return ok && e.Timeout()
	})
}

// IsNonRetryable will check if err or one of its causes is a error that cannot be retried.
func IsNonRetryable(err error) bool {
	type nonRetryabler interface {
		NonRetryable() bool
	}
	return isErrorPredicate(err, func(err error) bool {
		e, ok := err.(nonRetryabler)
		return ok && e.NonRetryable()
	})
}

// isErrorPredicate returns true if err or one of its causes returns true when
// passed to p.
func isErrorPredicate(err error, p func(err error) bool) bool {
	type causer interface {
		Cause() error
	}

	errs := []error{err}

	// We often use multierr.Error which doesn't implement causer
	if me, ok := err.(*multierror.Error); ok {
		errs = me.Errors
	}

	for _, err := range errs {
		for err != nil {
			if p(err) {
				return true
			}
			cause, ok := err.(causer)
			if !ok {
				break
			}
			err = cause.Cause()
		}
	}

	return false
}
