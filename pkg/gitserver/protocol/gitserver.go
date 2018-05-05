package protocol

import (
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
)

// ExecRequest is a request to execute a command inside a git repository.
type ExecRequest struct {
	Repo api.RepoURI `json:"repo"`

	// URL is the repository's Git remote URL. If the gitserver already has cloned the repository,
	// this field is optional (it will use the last-used Git remote URL). If the repository is not
	// cloned on the gitserver, the request will fail.
	URL string `json:"url,omitempty"`

	EnsureRevision string          `json:"ensureRevision"`
	Args           []string        `json:"args"`
	Opt            *vcs.RemoteOpts `json:"opt"`
}

// RepoUpdateRequest is a request to update the contents of a given repo, or clone it if it doesn't exist.
type RepoUpdateRequest struct {
	Repo api.RepoURI `json:"repo"`

	// URL is the repository's Git remote URL (from which to clone or update).
	URL string `json:"url"`
}

type NotFoundPayload struct {
	CloneInProgress bool `json:"cloneInProgress"` // If true, exec returned with noop because clone is in progress.

	// CloneProgress is a progress message from the running clone command.
	CloneProgress string `json:"cloneProgress,omitempty"`
}

// IsRepoCloneableRequest is a request to determine if a repo is cloneable.
type IsRepoCloneableRequest struct {
	// Repo is the repository to check.
	Repo api.RepoURI `json:"Repo"`

	// URL is the repository's Git remote URL.
	URL string `json:"url"`
}

// IsRepoCloneableResponse is the response type for the IsRepoCloneableRequest.
type IsRepoCloneableResponse struct {
	Cloneable bool   // whether the repo is cloneable
	Reason    string // if not cloneable, the reason why not
}

// IsRepoClonedRequest is a request to determine if a repo currently exists on gitserver.
type IsRepoClonedRequest struct {
	// Repo is the repository to check.
	Repo api.RepoURI
}

// RepoInfoRequest is a request for information about a repository on gitserver.
type RepoInfoRequest struct {
	// Repo is the repository to get information about.
	Repo api.RepoURI
}

// RepoInfoResponse is the response to a repository information request (RepoInfoRequest).
type RepoInfoResponse struct {
	URL             string     // this repository's Git remote URL
	CloneInProgress bool       // whether the repository is currently being cloned
	CloneProgress   string     // a progress message from the running clone command.
	Cloned          bool       // whether the repository has been cloned successfully
	LastFetched     *time.Time // when the last `git remote update` or `git fetch` occurred
}

// CreatePatchFromPatchRequest is the request information needed for creating
// the simulated staging area git object for a repo.
type CreatePatchFromPatchRequest struct {
	// Repo is the repository to get information about.
	Repo api.RepoURI
	// BaseCommit is the revision that the staging area object is based on
	BaseCommit api.CommitID
	// Patch is the diff contents to be used to create the staging area revision
	Patch string
	// TargetRef is the ref that will be created for this patch
	TargetRef string
	// CommitInfo is the information that will be used when creating the commit from a patch
	CommitInfo vcs.PatchCommitInfo
}

// CreatePatchFromPatchResponse is the response type returned after creating
// a staging object for Phabricator
type CreatePatchFromPatchResponse struct {
	// Rev is the tag that the staging object can be found at
	Rev string
}
