package errcode

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/schema"

	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

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

func (err *HTTPErr) HTTPStatusCode() int {
	return err.Status
}

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

func IsHTTPErrorCode(err error, statusCode int) bool {
	return HTTP(err) == statusCode
}
