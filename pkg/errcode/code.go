// Package errcode maps Go errors to HTTP status codes as well as other useful
// functions for inspecting errors.
package errcode

import (
	"fmt"
	"net/http"
	"os"

	"context"

	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

	"github.com/gorilla/schema"
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
	case vcs.ErrRevisionNotFound:
		return http.StatusNotFound
	case context.DeadlineExceeded:
		return http.StatusRequestTimeout
	}

	if (vcs.IsRepoNotExist(err) && err.(vcs.RepoNotExistError).CloneInProgress) || strings.Contains(err.Error(), vcs.RepoNotExistError{CloneInProgress: true}.Error()) {
		return http.StatusAccepted
	} else if (vcs.IsRepoNotExist(err) && !err.(vcs.RepoNotExistError).CloneInProgress) || strings.Contains(err.Error(), vcs.RepoNotExistError{}.Error()) {
		return http.StatusNotFound
	} else if err == vcs.ErrRepoExist {
		return http.StatusConflict
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

// IsNotFound will check if err or one of its causes is a not found error.
func IsNotFound(err error) bool {
	type notFounder interface {
		NotFound() bool
	}
	return isErrorPredicate(err, func(err error) bool {
		e, ok := err.(notFounder)
		return ok && e.NotFound()
	})
}

// isErrorPredicate returns true if err or one of its causes returns true when
// passed to p.
func isErrorPredicate(err error, p func(err error) bool) bool {
	type causer interface {
		Cause() error
	}

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
	return false
}
