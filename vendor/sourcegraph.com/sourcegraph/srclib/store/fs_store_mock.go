package store

import (
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// A RepoStoreImporter implements both RepoStore and RepoImporter.
type MockRepoStoreImporter struct {
	MockRepoStore
	Import_        func(commitID string, unit *unit.SourceUnit, data graph.Output) error
	CreateVersion_ func(commitID string) error
}

func (m MockRepoStoreImporter) Import(commitID string, unit *unit.SourceUnit, data graph.Output) error {
	return m.Import_(commitID, unit, data)
}

func (m MockRepoStoreImporter) CreateVersion(commitID string) error {
	return m.CreateVersion_(commitID)
}

var _ RepoStoreImporter = (*MockRepoStoreImporter)(nil)
