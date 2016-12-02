package store

import (
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// MultiRepoStore provides access to RepoStores for multiple
// repositories.
//
// Using this interface instead of directly accessing a single
// RepoStore allows aliasing repository URIs and supporting both ID
// and URI lookups.
type MultiRepoStore interface {
	// Repos returns all repositories that match the RepoFilter.
	Repos(...RepoFilter) ([]string, error)

	// RepoStore's methods call the corresponding methods on the
	// RepoStore of each repository contained within this multi-repo
	// store. The combined results are returned (in undefined order).
	RepoStore
}

// A MultiRepoImporter imports srclib build data for a repository's
// source unit at a specific version into a RepoStore.
type MultiRepoImporter interface {
	// Import imports srclib build data for a source unit at a
	// specific version into the store.
	Import(repo, commitID string, unit *unit.SourceUnit, data graph.Output) error

	// CreateVersion creates the version entry for the given commit. All other data (including
	// indexes) needs to exist before this gets called.
	CreateVersion(repo, commitID string) error
}

type MultiRepoIndexer interface {
	// Index builds indexes for the store.
	Index(repo, commitID string) error
}

// A MultiRepoStoreImporter implements both MultiRepoStore and
// MultiRepoImporter.
type MultiRepoStoreImporter interface {
	MultiRepoStore
	MultiRepoImporter
}

// A MultiRepoStoreImporterIndexer implements all 3 interfaces.
type MultiRepoStoreImporterIndexer interface {
	MultiRepoStore
	MultiRepoImporter
	MultiRepoIndexer
}

// TODO(sqs): What should the Repo type be? Right now it is just string.
