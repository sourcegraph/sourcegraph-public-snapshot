// Package gitserver contains a server that manages git repositories on disk,
// and a client that provides remote access to them.
package gitserver

import "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

type request struct {
	Exec   *execRequest
	Search *searchRequest
	Create *createRequest
	Remove *removeRequest
}

// execRequest ...
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

func (r *execReply) repoFound() bool { return !r.RepoNotFound }

type processResult struct {
	Error      string
	ExitStatus int
}

// searchRequest ...
type searchRequest struct {
	Repo      string
	Commit    vcs.CommitID
	Opt       vcs.SearchOptions
	ReplyChan chan<- *searchReply
}

type searchReply struct {
	RepoNotFound    bool // If true, search returned with noop because repo is not found.
	CloneInProgress bool // If true, search returned with noop because clone is in progress.
	Results         []*vcs.SearchResult
	Error           string // If non-empty, an error happened.
}

func (r *searchReply) repoFound() bool { return !r.RepoNotFound }

// createRequest ...
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

// removeRequest ...
type removeRequest struct {
	Repo      string
	ReplyChan chan<- *removeReply
}

type removeReply struct {
	RepoNotFound    bool   // If true, remove returned with noop because repo is not found.
	CloneInProgress bool   // If true, remove returned with noop because clone is in progress.
	Error           string // If non-empty, an error happened.
}

func (r *removeReply) repoFound() bool { return !r.RepoNotFound }
