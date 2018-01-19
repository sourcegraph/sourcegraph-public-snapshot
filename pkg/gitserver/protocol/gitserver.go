package protocol

import (
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

// ExecRequest is a request to execute a command inside a git repository.
type ExecRequest struct {
	Repo           string          `json:"repo"`
	EnsureRevision string          `json:"ensureRevision"`
	Args           []string        `json:"args"`
	Opt            *vcs.RemoteOpts `json:"opt"`
}

// RepoUpdateRequest is a request to update the contents of a given repo, or clone it if it doesn't exist.
type RepoUpdateRequest struct {
	Repo string `json:"repo"`
}

type NotFoundPayload struct {
	CloneInProgress bool `json:"cloneInProgress"` // If true, exec returned with noop because clone is in progress.
}

// IsRepoCloneableRequest is a request to determine if a repo is cloneable.
type IsRepoCloneableRequest struct {
	// Repo is the repository to check.
	Repo string
}

// IsRepoCloneableResponse is the response type for the IsRepoCloneableRequest.
type IsRepoCloneableResponse struct {
	Cloneable bool   // whether the repo is cloneable
	Reason    string // if not cloneable, the reason why not
}

// IsRepoClonedRequest is a request to determine if a repo currently exists on gitserver.
type IsRepoClonedRequest struct {
	// Repo is the repository to check.
	Repo string
}

// RepoInfoRequest is a request for information about a repository on gitserver.
type RepoInfoRequest struct {
	// Repo is the repository to get information about.
	Repo string
}

// RepoInfoResponse is the response to a repository information request (RepoInfoRequest).
type RepoInfoResponse struct {
	URL             string     // this repository's clone URL
	CloneInProgress bool       // whether the repository is currently being cloned
	Cloned          bool       // whether the repository has been cloned successfully
	LastFetched     *time.Time // when the last `git remote update` or `git fetch` occurred
}
