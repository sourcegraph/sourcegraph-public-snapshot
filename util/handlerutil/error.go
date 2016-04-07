package handlerutil

// NoVCSDataError may be returned when VCS data is not available for a requested
// resource.
type NoVCSDataError struct {
	RepoCommon      *RepoCommon
	CloneInProgress bool // If true, we should wait, not do anything, and try again later. It will finish cloning soon.
}

func (e *NoVCSDataError) Error() string {
	return "No VCS data found for " + e.RepoCommon.Repo.URI
}

// URLMovedError should be returned when a requested resource has moved to a new
// address.
type URLMovedError struct {
	NewURL string `json:"RedirectTo"`
}

func (e *URLMovedError) Error() string {
	return "URL moved to " + e.NewURL
}
