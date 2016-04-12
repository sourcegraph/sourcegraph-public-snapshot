package store

import (
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

type MockMultiRepoStore struct {
	Repos_    func(...RepoFilter) ([]string, error)
	Versions_ func(...VersionFilter) ([]*Version, error)
	Units_    func(...UnitFilter) ([]*unit.SourceUnit, error)
	Defs_     func(...DefFilter) ([]*graph.Def, error)
	Refs_     func(...RefFilter) ([]*graph.Ref, error)

	Import_        func(repo, commitID string, unit *unit.SourceUnit, data graph.Output) error
	Index_         func(repo, commitID string) error
	CreateVersion_ func(repo, commit string) error
}

func (m MockMultiRepoStore) Repos(f ...RepoFilter) ([]string, error) {
	return m.Repos_(f...)
}

func (m MockMultiRepoStore) Versions(f ...VersionFilter) ([]*Version, error) {
	return m.Versions_(f...)
}

func (m MockMultiRepoStore) Units(f ...UnitFilter) ([]*unit.SourceUnit, error) {
	return m.Units_(f...)
}

func (m MockMultiRepoStore) Defs(f ...DefFilter) ([]*graph.Def, error) {
	return m.Defs_(f...)
}

func (m MockMultiRepoStore) Refs(f ...RefFilter) ([]*graph.Ref, error) {
	return m.Refs_(f...)
}

func (m MockMultiRepoStore) Import(repo, commitID string, unit *unit.SourceUnit, data graph.Output) error {
	return m.Import_(repo, commitID, unit, data)
}

func (m MockMultiRepoStore) Index(repo, commitID string) error {
	return m.Index_(repo, commitID)
}

func (m MockMultiRepoStore) CreateVersion(repo, commitID string) error {
	return m.CreateVersion_(repo, commitID)
}

var _ MultiRepoStoreImporterIndexer = MockMultiRepoStore{}
