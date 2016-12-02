package store

type MockRepoStore struct {
	Versions_ func(...VersionFilter) ([]*Version, error)
	MockTreeStore
}

func (m MockRepoStore) Versions(f ...VersionFilter) ([]*Version, error) {
	return m.Versions_(f...)
}

var _ RepoStore = MockRepoStore{}
