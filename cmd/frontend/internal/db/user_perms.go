package db

// ErrRepoNotFound indicates that the repo does not exist or that the user has no access to that
// repo. Those two cases are not differentiated to avoid leaking repo existence information.
var ErrRepoNotFound = &repoNotFoundErr{}

type repoNotFoundErr struct{}

func (e *repoNotFoundErr) Error() string {
	return "repo not found"
}

func (e *repoNotFoundErr) NotFound() bool {
	return true
}
