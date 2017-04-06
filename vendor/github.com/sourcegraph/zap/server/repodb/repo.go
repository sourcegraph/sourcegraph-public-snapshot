package repodb

import (
	"github.com/sourcegraph/zap"
	"github.com/sourcegraph/zap/server/refdb"
)

// A Repo represents a repository on the server.
//
// It is guaranteed to be current if it is contained inside an
// OwnedRepo that is locked. Otherwise, its Path and RefDB fields
// remain current, but its Config may become stale.
type Repo struct {
	// Path is the repo's path, in the original case (i.e., not
	// canonically cased on case-insensitive file systems).
	//
	// The Path field must not be changed. There is currently no API
	// for renaming a repo.
	Path string

	// RefDB is the database of refs for this repo.
	RefDB *refdb.SyncRefDB

	// Config is the repository's configuration.
	//
	// Callers that access this field without holding the repo path
	// lock see a stale version of it that was deep-copied at the
	// instant the OwnedRepo was unlocked. Callers that don't hold the
	// repo path lock also should not modify this field, since
	// modifications will be ignored.
	Config zap.RepoConfiguration

	// Workspace is the workspace.Workspace interface value
	// representing this repository's workspace (or nil for bare
	// repos).
	//
	// Only the holder of the repo path lock for this repository can
	// safely access this field. After calling Unlock, this field's
	// value is not safe to use.
	//
	// The type is interface{} to simplify the import
	// graph.
	Workspace interface{}

	// WorkspaceRef is the name of the ref that represents the
	// contents of the active workspace (i.e., on disk). It is empty
	// for bare repos.
	//
	// Only the holder of the repo path lock for this repository can
	// safely access this field. After calling Unlock, this field's
	// value is not safe to use.
	//
	// TODO(sqs8): this does not account for when there is an active
	// Zap branch.
	WorkspaceRef string

	SendRefUpdateUpstream chan<- zap.RefUpdateUpstreamParams
}

// An OwnedRepo is a Repo that holds an exclusive lock in a repodb on
// its repo path until its Unlock method is called.
type OwnedRepo struct {
	*Repo // the repo (always non-nil)

	path   string // the repo path that yielded this OwnedRepo (constraint: Repo.Path == path)
	unlock func() // unlocks the repo path lock in the containing repodb
}

// Unlock unlocks the exclusive lock on r's repo path in its
// repodb. After calling Unlock, r.Repo is no longer guaranteed to be
// current.
func (r *OwnedRepo) Unlock() {
	if r == nil {
		panic("r == nil")
	}
	if r.unlock == nil {
		panic("Unlock was already called on repo " + r.path)
	}

	// Shallow-copy the Repo and deep-copy the config so that callers
	// can access the (albeit stale) config value.
	tmp := *r.Repo
	tmp.Config = tmp.Config.DeepCopy()
	r.Repo = &tmp

	r.unlock()
	r.unlock = nil
}
