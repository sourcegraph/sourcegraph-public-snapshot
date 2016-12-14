// Package gitserver contains a server that manages git repositories on disk,
// and a client that provides remote access to them.
package gitserver

import (
	"github.com/opentracing/opentracing-go"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

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

func (r *execReply) repoFound() bool { return !r.RepoNotFound }

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

// setSpanTags sets the relevant span tags on span for this request.
func setSpanTags(span opentracing.Span, r *request) {
	switch {
	case r.Exec != nil:
		span.SetTag("request", "Exec")
		span.SetTag("repo", r.Exec.Repo)
		span.SetTag("args", r.Exec.Args)
		span.SetTag("opt", r.Exec.Opt)
	case r.Create != nil:
		span.SetTag("request", "Create")
		span.SetTag("repo", r.Create.Repo)
		span.SetTag("MirrorRemote", r.Create.MirrorRemote)
		span.SetTag("opt", r.Create.Opt)
	default:
		span.SetTag("request", "unknown type")
	}
}
