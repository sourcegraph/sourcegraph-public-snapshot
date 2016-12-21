// Package gitserver contains a server that manages git repositories on disk,
// and a client that provides remote access to them.
package gitserver

import "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

type request struct {
	Exec   *execRequest
	Create *createRequest
}

// execRequest is a request to execute a command inside a git repository.
type execRequest struct {
	Repo      string
	Args      []string
	Opt       *vcs.RemoteOpts
	Stdin     <-chan []byte
	ReplyChan chan<- *execReply
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

// createRequest is a request to create a git repository.
type createRequest struct {
	Repo         string
	MirrorRemote string
	Opt          *vcs.RemoteOpts
	ReplyChan    chan<- *createReply
}

type createReply struct {
	RepoExist       bool   // If true, create returned with noop because repo exists.
	CloneInProgress bool   // If true, create returned with noop because clone is in progress.
	Error           string // If non-empty, an error happened.
}
