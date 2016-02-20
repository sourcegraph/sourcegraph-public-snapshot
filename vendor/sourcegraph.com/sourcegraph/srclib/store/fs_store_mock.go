package store

import (
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// A RepoStoreImporter implements both RepoStore and RepoImporter.
type MockRepoStoreImporter struct {
	MockRepoStore
	Import_ func(commitID string, unit *unit.SourceUnit, data graph.Output) error
}

func (m MockRepoStoreImporter) Import(commitID string, unit *unit.SourceUnit, data graph.Output) error {
	return m.Import_(commitID, unit, data)
}

var _ RepoStoreImporter = (*MockRepoStoreImporter)(nil)
