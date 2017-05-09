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

type ExecError struct {
	RepoNotFound    bool   `json:"repoNotFound"`    // If true, exec returned with noop because repo is not found.
	CloneInProgress bool   `json:"cloneInProgress"` // If true, exec returned with noop because clone is in progress.
	Error           string `json:"error"`
	ExitStatus      int    `json:"exitStatus"`
	Stderr          string `json:"stderr"`
}

type LegacyRequest struct {
	Exec *LegacyExecRequest
}

// LegacyExecRequest is a request to execute a command inside a git repository.
type LegacyExecRequest struct {
	Repo           string
	EnsureRevision string
	Args           []string
	Opt            *vcs.RemoteOpts
	Stdin          <-chan []byte // deprecated
	ReplyChan      chan<- *LegacyExecReply

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
	NoAutoUpdate bool
}

type LegacyExecReply struct {
	RepoNotFound    bool // If true, exec returned with noop because repo is not found.
	CloneInProgress bool // If true, exec returned with noop because clone is in progress.
	Stdout          <-chan []byte
	Stderr          <-chan []byte
	ProcessResult   <-chan *ProcessResult
}

type ProcessResult struct {
	Error      string
	ExitStatus int
}
