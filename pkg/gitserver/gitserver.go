// Package gitserver contains a server that manages git repositories on disk,
// and a client that provides remote access to them.
package gitserver

import "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

type request struct {
	Exec *execRequest
}

// execRequest is a request to execute a command inside a git repository.
type execRequest struct {
	Repo           string
	EnsureRevision string
	Args           []string
	Opt            *vcs.RemoteOpts
	Stdin          <-chan []byte
	ReplyChan      chan<- *execReply

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

type execReply struct {
	RepoNotFound    bool // If true, exec returned with noop because repo is not found.
	CloneInProgress bool // If true, exec returned with noop because clone is in progress.
	Stdout          <-chan []byte
	Stderr          <-chan []byte
	ProcessResult   <-chan *processResult
}

type processResult struct {
	Error      string
	ExitStatus int
}
