package errcode

import (
	"fmt"
	"net/http"
	"os"

	"context"

	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

	"github.com/gorilla/schema"
	"github.com/sourcegraph/go-github/github"
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
	case auth.ErrNoExternalAuthToken:
		return http.StatusUnauthorized
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
	case *github.ErrorResponse:
		return e.Response.StatusCode
	case schema.ConversionError:
		return http.StatusBadRequest
	case schema.MultiError:
		return http.StatusBadRequest
	case legacyerr.Error:
		return codeToHTTP(e.Code)
	}

	if os.IsNotExist(err) {
		return http.StatusNotFound
	} else if os.IsPermission(err) {
		return http.StatusForbidden
	}

	return http.StatusInternalServerError
}

// Code returns the most appropriate error code that describes
// err.
func Code(err error) legacyerr.Code {
	// Piggyback on the HTTP func to reduce code duplication.
	return HTTPToCode(HTTP(err))
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
