package mockstore

import (
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

func GraphMockDefs(m *store.MockMultiRepoStore, wantDefs ...*graph.Def) (called *bool) {
	called = new(bool)
	m.Defs_ = func(...store.DefFilter) ([]*graph.Def, error) {
		*called = true
		return wantDefs, nil
	}
	return
}

func GraphMockRefs(m *store.MockMultiRepoStore, wantRefs ...*graph.Ref) (called *bool) {
	called = new(bool)
	m.Refs_ = func(...store.RefFilter) ([]*graph.Ref, error) {
		*called = true
		return wantRefs, nil
	}
	return
}

func GraphMockUnits(m *store.MockMultiRepoStore, wantUnits ...*unit.SourceUnit) (called *bool) {
	called = new(bool)
	m.Units_ = func(...store.UnitFilter) ([]*unit.SourceUnit, error) {
		*called = true
		return wantUnits, nil
	}
	return
}

func GraphMockVersions(m *store.MockMultiRepoStore, wantVersions ...*store.Version) (called *bool) {
	called = new(bool)
	m.Versions_ = func(...store.VersionFilter) ([]*store.Version, error) {
		*called = true
		return wantVersions, nil
	}
	return
}

// GraphMockVersionsFiltered mocks m.Versions to make it behave as
// though its underlying storage contains all versions listed in
// wantVersions. When m.Versions is called with filters, those filters
// are applied to limit the results to only those versions that match
// the filters (unlike GraphMockVersions, which ignores the filters).
func GraphMockVersionsFiltered(m *store.MockMultiRepoStore, wantVersions ...*store.Version) (called *bool) {
	called = new(bool)
	m.Versions_ = func(fs ...store.VersionFilter) ([]*store.Version, error) {
		var vers []*store.Version
		for _, ver := range wantVersions {
			if versionFilters(fs).SelectVersion(ver) {
				vers = append(vers, ver)
			}
		}
		*called = true
		return vers, nil
	}
	return
}

type versionFilters []store.VersionFilter

func (fs versionFilters) SelectVersion(version *store.Version) bool {
	for _, f := range fs {
		if !f.SelectVersion(version) {
			return false
		}
	}
	return true
}
