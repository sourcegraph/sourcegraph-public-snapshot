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
