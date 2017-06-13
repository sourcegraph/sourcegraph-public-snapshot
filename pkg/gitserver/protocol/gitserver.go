package protocol

import "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

// ExecRequest is a request to execute a command inside a git repository.
type ExecRequest struct {
	Repo           string          `json:"repo"`
	EnsureRevision string          `json:"ensureRevision"`
	Args           []string        `json:"args"`
	Opt            *vcs.RemoteOpts `json:"opt"`

	// NoAutoUpdate is whether to prevent gitserver from auto-updating or cloning a repository if it
	// does not yet exist. This should be set to true if the following conditions hold:
	//
	// - the repo is private
	// - no auth credentials are provided
	//
	// Note: this addresses the following scenario:
	//
	// - unauthed client triggers gitserver to auto-clone a private repository without credentials
	// - authed client triggers auto-clone, which fails because there is an existing clone in progress
	// - unauthed auto-clone fails due to lack of credentials
	// - end result: nothing gets cloned, even though an authed client made a request that should have triggered an auto-clone
	//
	// There maybe a more elegant solution, but this will do for now.
	NoAutoUpdate bool `json:"noAutoUpdate"`
}

type NotFoundPayload struct {
	CloneInProgress bool `json:"cloneInProgress"` // If true, exec returned with noop because clone is in progress.
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
