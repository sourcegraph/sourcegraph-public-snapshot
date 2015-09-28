package store

import (
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// A TreeStore stores and accesses srclib build data for an arbitrary
// source tree (consisting of any number of source units).
type TreeStore interface {
	// Units returns all units that match the filter.
	Units(...UnitFilter) ([]*unit.SourceUnit, error)

	// UnitStore's methods call the corresponding methods on the
	// UnitStore of each source unit contained within this tree. The
	// combined results are returned (in undefined order).
	UnitStore
}

// A TreeImporter imports srclib build data for a source unit into a
// TreeStore.
type TreeImporter interface {
	// Import imports a source unit and its graph data into the
	// store. If Import is called with a nil SourceUnit and output
	// data, the importer considers the tree to have no source units
	// until others are imported in the future (this makes it possible
	// to distinguish between a tree that has no source units and a
	// tree whose source units simply haven't been imported yet).
	Import(*unit.SourceUnit, graph.Output) error
}

// A TreeStoreImporter implements both TreeStore and TreeImporter.
type TreeStoreImporter interface {
	TreeStore
	TreeImporter
}

type TreeIndexer interface {
	// Index builds indexes for the store, which may include data from
	// multiple source units in the tree.
	Index() error
}

// A treeStores is a TreeStore whose methods call the
// corresponding method on each of the tree stores returned by the
// treeStores func.
type treeStores struct {
	opener treeStoreOpener
}

var _ TreeStore = (*treeStores)(nil)

func (s treeStores) Units(f ...UnitFilter) ([]*unit.SourceUnit, error) {
	tss, err := openTreeStores(s.opener, f)
	if err != nil {
		return nil, err
	}

	var allUnits []*unit.SourceUnit
	for commitID, ts := range tss {
		if ts == nil {
			continue
		}

		units, err := ts.Units(f...)
		if err != nil && !isStoreNotExist(err) {
			return nil, err
		}
		for _, unit := range units {
			unit.CommitID = commitID
		}
		allUnits = append(allUnits, units...)
	}
	return allUnits, nil
}

func (s treeStores) Defs(f ...DefFilter) ([]*graph.Def, error) {
	tss, err := openTreeStores(s.opener, f)
	if err != nil {
		return nil, err
	}

	var allDefs []*graph.Def
	for commitID, ts := range tss {
		if ts == nil {
			continue
		}

		defs, err := ts.Defs(f...)
		if err != nil && !isStoreNotExist(err) {
			return nil, err
		}
		for _, def := range defs {
			def.CommitID = commitID
		}
		allDefs = append(allDefs, defs...)
	}
	return allDefs, nil
}

func (s treeStores) Refs(f ...RefFilter) ([]*graph.Ref, error) {
	tss, err := openTreeStores(s.opener, f)
	if err != nil {
		return nil, err
	}

	var allRefs []*graph.Ref
	for commitID, ts := range tss {
		if ts == nil {
			continue
		}

		setImpliedCommitID(f, commitID)
		refs, err := ts.Refs(f...)
		if err != nil && !isStoreNotExist(err) {
			return nil, err
		}
		for _, ref := range refs {
			ref.CommitID = commitID
		}
		allRefs = append(allRefs, refs...)
	}
	return allRefs, nil
}
