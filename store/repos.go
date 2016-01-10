package store

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/gitproto"
)

// Repos defines the interface for stores that persist and query
// repositories.
type Repos interface {
	// Get gets a repository.
	Get(ctx context.Context, repo string) (*sourcegraph.Repo, error)

	// GetPerms reports the current user's (or anonymous user's)
	// permissions on the specified repo.
	GetPerms(ctx context.Context, repo string) (*sourcegraph.RepoPermissions, error)

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
//
// Note: See the RepoOrigins doc for more information on the split
// between Sourcegraph-specific data and origin-specific data.
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
	Clone(ctx context.Context, repo string, bare, mirror bool, info *vcsclient.CloneInfo) error
	OpenGitTransport(ctx context.Context, repo string) (gitproto.Transport, error)
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

// ErrRepoNeedsCloneURL occurs when Repos.Create is called and the
// repo has no HTTPCloneURL or SSHCloneURL set, when the store type
// requires that one be set. For example, the DB store requires the
// repo to exist as a git/hg/etc. repository elsewhere already, since
// it only creates a row in the DB for the repo, not the actual VCS
// repository.
var ErrRepoNeedsCloneURL = errors.New("creating a repo requires a clone URL to be set")

// ErrRepoNoCloneURL occurs when Repos.Create is called and the repo
// has a HTTPCloneURL or SSHCloneURL set, when the store type requires
// that one NOT be set. For example, the FS store calls `git init` to
// create a new repo, so a clone URL would be meaningless.
var ErrRepoNoCloneURL = errors.New("creating a hosted repo initializes a new repo; no clone URL may be provided")

// ErrRepoMirrorOnly occurs when Repos.Create is called to create a
// non-mirror repo, but the store type requires mirror repos (e.g.,
// the DB-backed store).
var ErrRepoMirrorOnly = errors.New("this repo store can only create mirrored repos")
