package gitcmd

import (
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type Repository struct {
	// repoURI is the identifier of the repository, which is opaque to gitserver and is
	// conventionally a string like "github.com/gorilla/mux".
	repoURI api.RepoURI

	// once protects runs of remoteURLFunc and writes to remoteURL.
	once sync.Once

	// remoteURLFunc will be run to set the remoteURL.
	remoteURLFunc func() (string, error)

	// remoteURL and remoteURLErr is the output of remoteURLFunc.
	remoteURL    string
	remoteURLErr error
}

func (r *Repository) String() string {
	return fmt.Sprintf("git repo %s", r.repoURI)
}

// Open returns a handle to a repository on gitserver with the given
// identifier (repoURI) and optional Git remote URL. The Git remote URL is
// only required if the gitserver doesn't already contain a clone of the
// repository or if the revision must be fetched from the remote. This only
// happens when calling ResolveRevision.
//
// TODO(sqs!): move to gitserver client?
func Open(repoURI api.RepoURI, remoteURL string) *Repository {
	return &Repository{repoURI: repoURI, remoteURLFunc: func() (string, error) {
		return remoteURL, nil
	}}
}

// OpenLazy returns a handle to a repository on gitserver with the given
// identifier (repoURI) and optional Git remote URL function. The Git remote
// URL is only required if the gitserver doesn't already contain a clone of
// the repository or if the revision must be fetched from the remote. This
// only happens when calling ResolveRevision.
func OpenLazy(repoURI api.RepoURI, remoteURLFunc func() (string, error)) *Repository {
	return &Repository{repoURI: repoURI, remoteURLFunc: remoteURLFunc}
}
