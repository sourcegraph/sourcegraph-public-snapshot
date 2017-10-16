package protocol

import "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

// ExecRequest is a request to execute a command inside a git repository.
type ExecRequest struct {
	Repo           string          `json:"repo"`
	EnsureRevision string          `json:"ensureRevision"`
	Args           []string        `json:"args"`
	Opt            *vcs.RemoteOpts `json:"opt"`
}

type NotFoundPayload struct {
	CloneInProgress bool `json:"cloneInProgress"` // If true, exec returned with noop because clone is in progress.
}

// IsRepoCloneableRequest is a request to determine if a repo is cloneable.
type IsRepoCloneableRequest struct {
	// Repo is the repository to check.
	Repo string
}

// RepoFromRemoteURLRequest is a request to determine a repository URI (like
// github.com/gorilla/mux) from a Git remote URL (like git@github.com:gorilla/mux
// or any other variation).
type RepoFromRemoteURLRequest struct {
	// RemoteURL is the remote URL to derive a repository URI from. It may be
	// of any valid Git form, and may come from user input (in which case, it
	// may contain private credentials if they use basic auth).
	RemoteURL string
}
