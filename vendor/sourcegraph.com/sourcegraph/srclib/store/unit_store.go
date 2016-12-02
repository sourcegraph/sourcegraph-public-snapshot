package store

import (
	"sync"

	"github.com/neelance/parallel"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

// A UnitStore stores and accesses srclib build data for a single
// source unit.
type UnitStore interface {
	// Defs returns all defs that match the filter.
	Defs(...DefFilter) ([]*graph.Def, error)

	// Refs returns all refs that match the filter.
	Refs(...RefFilter) ([]*graph.Ref, error)

	// TODO(sqs): how to deal with depresolve and other non-graph
	// data?
}

// A UnitImporter imports srclib build data for a single source unit
// into a UnitStore.
type UnitImporter interface {
	// Import imports defs, refs, etc., into the store. It overwrites
	// all existing data for this source unit (and at the commit, if
	// applicable).
	Import(graph.Output) error
}

// A UnitStoreImporter implements both UnitStore and UnitImporter.
type UnitStoreImporter interface {
	UnitStore
	UnitImporter
}

// A unitStores is a UnitStore whose methods call the
// corresponding method on each of the unit stores returned by the
// unitStores func.
type unitStores struct {
	opener unitStoreOpener
}

var _ UnitStore = (*unitStores)(nil)

func (s unitStores) Defs(fs ...DefFilter) ([]*graph.Def, error) {
	uss, err := openUnitStores(s.opener, fs)
	if err != nil {
		return nil, err
	}

	var (
		allDefs   []*graph.Def
		allDefsMu sync.Mutex
	)
	par := parallel.NewRun(storeFetchPar)
	for u_, us_ := range uss {
		u, us := u_, us_
		if us == nil {
			continue
		}

		par.Acquire()
		go func() {
			defer par.Release()
			defs, err := us.Defs(filtersForUnit(u, fs).([]DefFilter)...)
			if err != nil && !isStoreNotExist(err) {
				par.Error(err)
				return
			}
			for _, def := range defs {
				def.UnitType = u.Type
				def.Unit = u.Name
			}
			allDefsMu.Lock()
			allDefs = append(allDefs, defs...)
			allDefsMu.Unlock()
		}()
	}
	err = par.Wait()
	return allDefs, err
}

var c_unitStores_Refs_last_numUnitsQueried = &counter{count: new(int64)}

func (s unitStores) Refs(f ...RefFilter) ([]*graph.Ref, error) {
	uss, err := openUnitStores(s.opener, f)
	if err != nil {
		return nil, err
	}

	c_unitStores_Refs_last_numUnitsQueried.set(0)
	var (
		allRefsMu sync.Mutex
		allRefs   []*graph.Ref
	)
	par := parallel.NewRun(storeFetchPar)
	for u, us := range uss {
		if us == nil {
			continue
		}
		u, us := u, us

		c_unitStores_Refs_last_numUnitsQueried.increment()

		par.Acquire()
		go func() {
			defer par.Release()
			if _, moreOK := LimitRemaining(f); !moreOK {
				return
			}
			fCopy := filtersForUnit(u, f).([]RefFilter)
			fCopy = withImpliedUnit(fCopy, u)

			refs, err := us.Refs(fCopy...)
			if err != nil && !isStoreNotExist(err) {
				par.Error(err)
				return
			}
			for _, ref := range refs {
				ref.UnitType = u.Type
				ref.Unit = u.Name
				if ref.DefUnitType == "" {
					ref.DefUnitType = u.Type
				}
				if ref.DefUnit == "" {
					ref.DefUnit = u.Name
				}
			}

			allRefsMu.Lock()
			allRefs = append(allRefs, refs...)
			allRefsMu.Unlock()
		}()
	}
	err = par.Wait()
	return allRefs, err
}

func cleanForImport(data *graph.Output, repo, unitType, unit string) {
	for _, def := range data.Defs {
		def.Unit = ""
		def.UnitType = ""
		def.Repo = ""
		def.CommitID = ""
	}
	for _, ref := range data.Refs {
		ref.Unit = ""
		ref.UnitType = ""
		ref.Repo = ""
		ref.CommitID = ""
		if repo != "" && ref.DefRepo == repo {
			ref.DefRepo = ""
		}
		if unitType != "" && ref.DefUnitType == unitType {
			ref.DefUnitType = ""
		}
		if unit != "" && ref.DefUnit == unit {
			ref.DefUnit = ""
		}
	}
	for _, doc := range data.Docs {
		doc.Unit = ""
		doc.UnitType = ""
		doc.Repo = ""
		doc.CommitID = ""
	}
	for _, ann := range data.Anns {
		ann.Unit = ""
		ann.UnitType = ""
		ann.Repo = ""
		ann.CommitID = ""
	}
}
