package repoupdater

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// ErrNotFound is an error that occurs when a repo doesn't exist.
type ErrNotFound struct {
	repo     api.RepoName
	notFound bool
}

// ErrNotFound returns true if the repo does not exist.
func (e *ErrNotFound) NotFound() bool {
	return e.notFound
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("repo not found (name=%s  notfound=%v) ", e.repo, e.notFound)
}

// ErrUnauthorized is an error that occurs when repo access is
// unauthorized.
type ErrUnauthorized struct {
	repo    api.RepoName
	noAuthz bool
}

// Unauthorized returns true if repo access is unauthorized.
func (e *ErrUnauthorized) Unauthorized() bool {
	return e.noAuthz
}

func (e *ErrUnauthorized) Error() string {
	return fmt.Sprintf("repo access not authorized (name=%s  noauthz=%v) ", e.repo, e.noAuthz)
}

// ErrTemporary is an error that can be retried
type ErrTemporary struct {
	repo        api.RepoName
	isTemporary bool
}

// ErrTemporary is when the repository was reported as being temporarily
// unavailable.
func (e *ErrTemporary) Temporary() bool {
	return e.isTemporary
}

func (e *ErrTemporary) Error() string {
	return fmt.Sprintf("repo temporary unavailable (name=%s  isTemporary=%v) ", e.repo, e.isTemporary)
}
