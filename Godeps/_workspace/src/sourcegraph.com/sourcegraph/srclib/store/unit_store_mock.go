package store

import "sourcegraph.com/sourcegraph/srclib/graph"

type MockUnitStore struct {
	Defs_ func(...DefFilter) ([]*graph.Def, error)
	Refs_ func(...RefFilter) ([]*graph.Ref, error)
}

func (m MockUnitStore) Defs(f ...DefFilter) ([]*graph.Def, error) {
	return m.Defs_(f...)
}

func (m MockUnitStore) Refs(f ...RefFilter) ([]*graph.Ref, error) {
	return m.Refs_(f...)
}

var _ UnitStore = MockUnitStore{}
