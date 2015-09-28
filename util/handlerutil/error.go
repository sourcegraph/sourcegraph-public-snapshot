package handlerutil

import (
	"fmt"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/errcode"
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

func (err *HTTPErr) HTTPStatusCode() int { return err.Status }

func IsHTTPErrorCode(err error, statusCode int) bool {
	return errcode.HTTP(err) == statusCode
}

// NoBuildError is returned whenever a build is requested for an unbuilt repo.
type NoBuildError struct {
	RepoCommon
	RepoRevCommon
	NeedsLogin bool
	Build      *sourcegraph.Build
}

func (e *NoBuildError) Error() string {
	spec := e.RepoRevCommon.RepoRevSpec
	return fmt.Sprintf("No build for repo %s@%s", spec.URI, spec.CommitID)
}

// NoVCSDataError may be returned when VCS data is not available for a requested
// resource.
type NoVCSDataError struct {
	RepoCommon *RepoCommon
}

func (e *NoVCSDataError) Error() string {
	return "No VCS data found for " + e.RepoCommon.Repo.URI
}

// RepoNotEnabledError may be returned when the requested repository has not yet
// been enabled on Sourcegraph.
type RepoNotEnabledError struct {
	RepoCommon *RepoCommon
}

func (e *RepoNotEnabledError) Error() string {
	return "Repo " + e.RepoCommon.Repo.URI + " not enabled."
}

// URLMovedError should be returned when a requested resource has moved to a new
// address.
type URLMovedError struct {
	NewURL string `json:"RedirectTo"`
}

func (e *URLMovedError) Error() string {
	return "URL moved to " + e.NewURL
}
