package store

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitproto"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

// Repos defines the interface for stores that persist and query
// repositories.
type Repos interface {
	// Get a repository.
	Get(ctx context.Context, repo string) (*sourcegraph.Repo, error)

	// List repositories.
	List(context.Context, *sourcegraph.RepoListOptions) ([]*sourcegraph.Repo, error)

	// Create a repository.
	Create(context.Context, *sourcegraph.Repo) error

	// Update a repository.
	Update(context.Context, RepoUpdate) error

	// Delete a repository.
	Delete(ctx context.Context, repo string) error
}

// RepoUpdate represents an update to specific fields of a repo. Only
// fields with non-zero values are updated.
//
// The ReposUpdateOp.Repo field must be filled in to specify the repo
// that will be updated.
type RepoUpdate struct {
	*sourcegraph.ReposUpdateOp

	UpdatedAt *time.Time
	PushedAt  *time.Time
}

// RepoConfigs is the interface for storing Sourcegraph-specific repo
// config.
type RepoConfigs interface {
	Get(ctx context.Context, repo string) (*sourcegraph.RepoConfig, error)
	Update(ctx context.Context, repo string, conf sourcegraph.RepoConfig) error
}

// RepoStatuses defines the interface for stores that deal with
// per-commit status message.
type RepoStatuses interface {
	GetCombined(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (*sourcegraph.CombinedStatus, error)
	GetCoverage(ctx context.Context) (*sourcegraph.RepoStatusList, error)
	Create(ctx context.Context, repoRev sourcegraph.RepoRevSpec, status *sourcegraph.RepoStatus) error
}

type RepoVCS interface {
	Open(ctx context.Context, repo string) (vcs.Repository, error)
	Clone(ctx context.Context, repo string, info *CloneInfo) error
	OpenGitTransport(ctx context.Context, repo string) (gitproto.Transport, error)
}

// CloneInfo is the information needed to clone a repository.
type CloneInfo struct {
	// VCS is the type of VCS (e.g., "git")
	VCS string
	// CloneURL is the remote URL from which to clone.
	CloneURL string
	// Additional options
	vcs.RemoteOpts
}

// RepoNotFoundError occurs when a repository is not found.
type RepoNotFoundError struct {
	Repo string // the requested repo
}

func (e *RepoNotFoundError) Error() string { return fmt.Sprintf("repo %s not found", e.Repo) }

// IsRepoNotFound returns true iff err is a *RepoNotFoundError.
func IsRepoNotFound(err error) bool {
	_, ok := err.(*RepoNotFoundError)
	return ok
}

// RepoExistError occurs when a repository already exists (but wasn't expected to).
type RepoExistError struct {
	URI string // the requested repo
}

func (e *RepoExistError) Error() string { return fmt.Sprintf("repo %q already exists", e.URI) }

// IsRepoExist returns true iff err is a *RepoExistError.
func IsRepoExist(err error) bool {
	_, ok := err.(*RepoExistError)
	return ok
}
