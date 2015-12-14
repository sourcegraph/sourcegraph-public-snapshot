package handlerutil

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
