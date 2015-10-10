package errcode

import (
	"net/http"
	"os"

	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/fed/discover"
	"src.sourcegraph.com/sourcegraph/store"

	"github.com/gorilla/schema"
	"github.com/sourcegraph/go-github/github"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
	case graph.ErrDefNotExist, sourcegraph.ErrBuildNotFound, vcs.ErrRevisionNotFound, vcs.ErrCommitNotFound:
		return http.StatusNotFound
	case auth.ErrNoExternalAuthToken:
		return http.StatusUnauthorized
	case store.ErrRepoNeedsCloneURL, store.ErrRepoNoCloneURL, store.ErrRepoMirrorOnly:
		return http.StatusPreconditionFailed
	case store.ErrRegisteredClientIDExists:
		return http.StatusConflict
	}

	if strings.Contains(err.Error(), "git repository not found") {
		return http.StatusNotFound
	}

	switch e := err.(type) {
	case *sourcegraph.NotImplementedError:
		// Ignore NotImplementedError's HTTPStatusCode method (which
		// returns 404).
		return http.StatusNotImplemented
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
	case *store.RepoNotFoundError:
		return http.StatusNotFound
	case *store.UserNotFoundError:
		return http.StatusNotFound
	case *store.RegisteredClientNotFoundError:
		return http.StatusNotFound
	case *store.AccountAlreadyExistsError:
		return http.StatusConflict
	}

	if os.IsNotExist(err) {
		return http.StatusNotFound
	} else if os.IsNotExist(err) {
		return http.StatusNotFound
	} else if os.IsPermission(err) {
		return http.StatusForbidden
	} else if discover.IsNotFound(err) {
		return http.StatusNotFound
	}

	if code := grpc.Code(err); code != codes.Unknown {
		return grpcToHTTP(code)
	}

	return http.StatusInternalServerError
}

// GRPC returns the most appropriate gRPC error code that describes
// err.
func GRPC(err error) codes.Code {
	// Piggyback on the HTTP func to reduce code duplication.
	return httpToGRPC(HTTP(err))
}
