package store

import (
	"errors"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// A memoryMultiRepoStore is a MultiRepoStore that stores data in
// memory
type memoryMultiRepoStore struct {
	repos map[string]*memoryRepoStore

	repoStores
}

func newMemoryMultiRepoStore() *memoryMultiRepoStore {
	mrs := &memoryMultiRepoStore{}
	mrs.repoStores = repoStores{mrs}
	return mrs
}

var errMultiRepoStoreNoInit = errors.New("multi-repo store not yet initialized")

func (s *memoryMultiRepoStore) Repos(f ...RepoFilter) ([]string, error) {
	if s.repos == nil {
		return nil, errMultiRepoStoreNoInit
	}

	repos := []string{}
	for repo := range s.repos {
		if repoFilters(f).SelectRepo(repo) {
			repos = append(repos, repo)
		}
	}
	return repos, nil
}

func (s *memoryMultiRepoStore) openRepoStore(repo string) RepoStore {
	if rs, present := s.repos[repo]; present {
		return rs
	}
	return nil
}

func (s *memoryMultiRepoStore) openAllRepoStores() (map[string]RepoStore, error) {
	if s.repos == nil {
		return nil, errMultiRepoStoreNoInit
	}

	rss := make(map[string]RepoStore, len(s.repos))
	for repo := range s.repos {
		rss[repo] = s.openRepoStore(repo)
	}
	return rss, nil
}

var _ repoStoreOpener = (*memoryMultiRepoStore)(nil)

func (s *memoryMultiRepoStore) Import(repo, commitID string, unit *unit.SourceUnit, data graph.Output) error {
	if s.repos == nil {
		s.repos = map[string]*memoryRepoStore{}
	}
	if _, present := s.repos[repo]; !present {
		s.repos[repo] = newMemoryRepoStore()
	}
	if unit != nil {
		cleanForImport(&data, repo, unit.Type, unit.Name)
	}
	return s.repos[repo].Import(commitID, unit, data)
}

func (s *memoryMultiRepoStore) CreateVersion(repo, commitID string) error {
	return s.repos[repo].CreateVersion(commitID)
}

func (s *memoryMultiRepoStore) String() string { return "memoryMultiRepoStore" }

// A memoryRepoStore is a RepoStore that stores data in memory.
type memoryRepoStore struct {
	versions []*Version
	trees    map[string]*memoryTreeStore
	treeStores
}

func newMemoryRepoStore() *memoryRepoStore {
	rs := &memoryRepoStore{}
	rs.treeStores = treeStores{rs}
	return rs
}

var errRepoNoInit = errors.New("repo not yet initialized")

func (s *memoryRepoStore) Versions(f ...VersionFilter) ([]*Version, error) {
	if s.versions == nil {
		return nil, errRepoNoInit
	}

	var versions []*Version
	for _, version := range s.versions {
		if versionFilters(f).SelectVersion(version) {
			versions = append(versions, version)
		}

	}
	return versions, nil
}

func (s *memoryRepoStore) Import(commitID string, unit *unit.SourceUnit, data graph.Output) error {
	if s.trees == nil {
		s.trees = map[string]*memoryTreeStore{}
	}
	if _, present := s.trees[commitID]; !present {
		s.trees[commitID] = newMemoryTreeStore()
	}
	if unit != nil {
		cleanForImport(&data, "", unit.Type, unit.Name)
	}
	return s.trees[commitID].Import(unit, data)
}

func (s *memoryRepoStore) CreateVersion(commitID string) error {
	s.versions = append(s.versions, &Version{CommitID: commitID})
	return nil
}

func (s *memoryRepoStore) openTreeStore(commitID string) TreeStore {
	if ts, present := s.trees[commitID]; present {
		return ts
	}
	return nil
}

func (s *memoryRepoStore) openAllTreeStores() (map[string]TreeStore, error) {
	if s.trees == nil {
		return nil, errRepoNoInit
	}

	tss := make(map[string]TreeStore, len(s.trees))
	for commitID := range s.trees {
		tss[commitID] = s.openTreeStore(commitID)
	}
	return tss, nil
}

var _ treeStoreOpener = (*memoryRepoStore)(nil)

func (s *memoryRepoStore) String() string { return "memoryRepoStore" }

// A memoryTreeStore is a TreeStore that stores data in memory.
type memoryTreeStore struct {
	units []*unit.SourceUnit
	data  map[unit.ID2]*graph.Output
	unitStores
}

func newMemoryTreeStore() *memoryTreeStore {
	ts := &memoryTreeStore{}
	ts.unitStores = unitStores{ts}
	return ts
}

var errTreeNoInit = errors.New("tree not yet initialized")

func (s *memoryTreeStore) Units(f ...UnitFilter) ([]*unit.SourceUnit, error) {
	if s.units == nil {
		return nil, errTreeNoInit
	}

	var units []*unit.SourceUnit
	for _, unit := range s.units {
		if unitFilters(f).SelectUnit(unit) {
			units = append(units, unit)
		}

	}
	return units, nil
}

func (s *memoryTreeStore) Import(u *unit.SourceUnit, data graph.Output) error {
	if s.units == nil {
		s.units = []*unit.SourceUnit{}
	}
	if s.data == nil {
		s.data = map[unit.ID2]*graph.Output{}
	}
	if u == nil {
		return nil
	}

	cleanForImport(&data, "", u.Type, u.Name)

	s.units = append(s.units, u)
	unitID := unit.ID2{Type: u.Type, Name: u.Name}
	s.data[unitID] = &data
	return nil
}

func (s *memoryTreeStore) openUnitStore(u unit.ID2) UnitStore {
	if dat, present := s.data[u]; present {
		return &memoryUnitStore{data: dat}
	}
	return nil
}

func (s *memoryTreeStore) openAllUnitStores() (map[unit.ID2]UnitStore, error) {
	if s.data == nil {
		return nil, errTreeNoInit
	}

	uss := make(map[unit.ID2]UnitStore, len(s.data))
	for u := range s.data {
		uss[u] = s.openUnitStore(u)
	}
	return uss, nil
}

var _ unitStoreOpener = (*memoryTreeStore)(nil)

func (s *memoryTreeStore) String() string { return "memoryTreeStore" }

// A memoryUnitStore is a UnitStore that stores data in memory.
type memoryUnitStore struct {
	data *graph.Output
}

var errUnitNoInit = errors.New("unit not yet initialized")

func (s *memoryUnitStore) Defs(f ...DefFilter) ([]*graph.Def, error) {
	if s.data == nil {
		return nil, errUnitNoInit
	}

	var defs []*graph.Def
	for _, def := range s.data.Defs {
		if DefFilters(f).SelectDef(def) {
			defs = append(defs, def)
		}
	}
	for _, filter := range f {
		if dSort, ok := filter.(DefsSorter); ok {
			dSort.DefsSort(defs)
			break
		}
	}
	return defs, nil
}

func (s *memoryUnitStore) Refs(f ...RefFilter) ([]*graph.Ref, error) {
	if s.data == nil {
		return nil, errUnitNoInit
	}

	var refs []*graph.Ref
	for _, ref := range s.data.Refs {
		if refFilters(f).SelectRef(ref) {
			refs = append(refs, ref)
		}
	}
	return refs, nil
}

func (s *memoryUnitStore) Import(data graph.Output) error {
	cleanForImport(&data, "", "", "")
	s.data = &data
	return nil
}

func (s *memoryUnitStore) String() string { return "memoryUnitStore" }
