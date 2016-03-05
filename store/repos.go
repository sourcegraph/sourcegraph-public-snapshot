package store

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/gitproto"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
)

// Repos defines the interface for stores that persist and query
// repositories.
type Repos interface {
	// Get gets a repository.
	Get(ctx context.Context, repo string) (*sourcegraph.Repo, error)

	// List lists repositories.
	List(context.Context, *sourcegraph.RepoListOptions) ([]*sourcegraph.Repo, error)

	// Create creates a repository.
	Create(context.Context, *sourcegraph.Repo) error

	// Update updates a repository.
	Update(context.Context, *RepoUpdate) error

	// Delete deletes a repository.
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
	Create(ctx context.Context, repoRev sourcegraph.RepoRevSpec, status *sourcegraph.RepoStatus) error
}

type RepoCounters interface {
	RecordHit(ctx context.Context, repo string) error
	CountHits(ctx context.Context, repo string, since time.Time) (int, error)
}

type RepoVCS interface {
	Open(ctx context.Context, repo string) (vcs.Repository, error)
	Clone(ctx context.Context, repo string, bare, mirror bool, info *CloneInfo) error
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
