package store

import (
	"time"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

// Repos defines the interface for stores that persist and query
// repositories.
type Repos interface {
	// Get a repository.
	Get(ctx context.Context, repo int32) (*sourcegraph.Repo, error)

	// GetByURI a repository by its URI.
	GetByURI(ctx context.Context, repo string) (*sourcegraph.Repo, error)

	// List repositories in the Sourcegraph repository store. Note:
	// this will not return any repositories from external services
	// that are not present in the Sourcegraph repository store.
	List(context.Context, *RepoListOp) ([]*sourcegraph.Repo, error)

	// Search repositories.
	//
	// DEPRECATED
	Search(context.Context, string) ([]*sourcegraph.RepoSearchResult, error)

	// Create a repository and return its ID.
	Create(context.Context, *sourcegraph.Repo) (int32, error)

	// Update a repository.
	Update(context.Context, RepoUpdate) error

	// InternalUpdate performs an update of internal repository
	// fields. See InternalRepoUpdate for more information.
	InternalUpdate(ctx context.Context, repo int32, op InternalRepoUpdate) error

	// Delete a repository.
	Delete(ctx context.Context, repo int32) error
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

type RepoListOp struct {
	// Name filters the repository list by name.
	Name string

	// Query specifies a search query for repositories. If specified, then the Sort and
	// Direction options are ignored
	Query string

	// URIs specifies a set of repository URIs to list.
	URIs []string

	// Sort determines how the returned list of repositories is sorted.
	Sort string

	// Direction determines the sort direction.
	Direction string

	// NoFork excludes forks from the list of returned repositories.
	NoFork bool

	// Type of repositories to list. Possible values are currently
	// ones supported by the GitHub API, including: all, owner,
	// public, private, member. Default is "all".
	Type string

	// Owner filters the list of repositories to those with the
	// specified owner.
	Owner string

	sourcegraph.ListOptions
}

// InternalRepoUpdate is an update of repo fields that are used by
// internal Sourcegraph processes only. It is separate from RepoUpdate
// so that internal updates can be performed by machine processes that
// do not need to assume the privileges of the repo's owner (and can
// merely have an internal token scope).
type InternalRepoUpdate struct {
	VCSSyncedAt *time.Time
}

// RepoConfigs is the interface for storing Sourcegraph-specific repo
// config.
type RepoConfigs interface {
	Get(ctx context.Context, repo int32) (*sourcegraph.RepoConfig, error)
	Update(ctx context.Context, repo int32, conf sourcegraph.RepoConfig) error
}

// RepoStatuses defines the interface for stores that deal with
// per-commit status message.
type RepoStatuses interface {
	GetCombined(ctx context.Context, repo int32, commitID string) (*sourcegraph.CombinedStatus, error)
	GetCoverage(ctx context.Context) (*sourcegraph.RepoStatusList, error)
	Create(ctx context.Context, repo int32, commitID string, status *sourcegraph.RepoStatus) error
}

type RepoVCS interface {
	Open(ctx context.Context, repo int32) (vcs.Repository, error)
	Clone(ctx context.Context, repo int32, info *CloneInfo) error
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
