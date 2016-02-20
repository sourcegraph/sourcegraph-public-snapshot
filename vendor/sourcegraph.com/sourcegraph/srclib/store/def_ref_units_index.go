package store

import (
	"fmt"
	"io"
	"sync"

	"github.com/alecthomas/binary"
	"github.com/gogo/protobuf/proto"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/store/phtable"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// NOTE(sqs): There is a lot of duplication here with unitFilesIndex.

// defRefUnitsIndex makes it fast to determine which source units
// contain refs to a def.
type defRefUnitsIndex struct {
	phtable *phtable.CHD
	ready   bool
	sync.RWMutex
}

var _ interface {
	Index
	persistedIndex
	unitRefIndexBuilder
	unitIndex
} = (*defRefUnitsIndex)(nil)

var c_defRefUnitsIndex_getByDef = &counter{count: new(int64)}

func (x *defRefUnitsIndex) String() string { return fmt.Sprintf("defRefUnitsIndex(ready=%v)", x.ready) }

// getByFile returns a list of source units that contain refs to the
// specified def.
func (x *defRefUnitsIndex) getByDef(def graph.RefDefKey) ([]unit.ID2, bool, error) {
	vlog.Printf("defRefUnitsIndex.getByDef(%v)", def)
	c_defRefUnitsIndex_getByDef.increment()

	k, err := proto.Marshal(&def)
	if err != nil {
		return nil, false, err
	}

	if x.phtable == nil {
		panic("phtable not built/read")
	}
	v := x.phtable.Get(k)
	if v == nil {
		return nil, false, nil
	}

	var us []unit.ID2
	if err := binary.Unmarshal(v, &us); err != nil {
		return nil, true, err
	}
	return us, true, nil
}

// Covers implements unitIndex.
func (x *defRefUnitsIndex) Covers(filters interface{}) int {
	cov := 0
	for _, f := range storeFilters(filters) {
		if _, ok := f.(ByRefDefFilter); ok {
			cov++
		}
	}
	return cov
}

// Units implements unitIndex.
func (x *defRefUnitsIndex) Units(fs ...UnitFilter) ([]unit.ID2, error) {
	x.RLock()
	defer x.RUnlock()
	for _, f := range fs {
		if ff, ok := f.(ByRefDefFilter); ok {
			us, found, err := x.getByDef(ff.withEmptyImpliedValues())
			if err != nil {
				return nil, err
			}
			if found {
				vlog.Printf("defRefUnitsIndex(%v): Found units %v using index.", fs, us)
				return us, nil
			}
		}
	}
	return nil, nil
}

// Build implements unitRefIndexBuilder.
func (x *defRefUnitsIndex) Build(unitRefIndexes map[unit.ID2]*defRefsIndex) error {
	x.Lock()
	defer x.Unlock()
	vlog.Printf("defRefUnitsIndex: building inverted def->units index (%d units)...", len(unitRefIndexes))
	defToUnits := map[graph.RefDefKey][]unit.ID2{}
	for u, x := range unitRefIndexes {
		it := x.phtable.Iterate()
		for {
			if it == nil {
				break
			}

			kb, _ := it.Get()
			var def graph.RefDefKey
			if err := proto.Unmarshal(kb, &def); err != nil {
				return err
			}

			// Set implied fields.
			if def.DefUnit == "" {
				def.DefUnit = u.Name
			}
			if def.DefUnitType == "" {
				def.DefUnitType = u.Type
			}
			defToUnits[def] = append(defToUnits[def], u)

			it = it.Next()
		}
	}
	vlog.Printf("defRefUnitsIndex: adding %d index phtable keys...", len(defToUnits))
	b := phtable.Builder(len(defToUnits))
	for def, units := range defToUnits {
		ub, err := binary.Marshal(units)
		if err != nil {
			return err
		}
		kb, err := proto.Marshal(&def)
		if err != nil {
			return err
		}
		b.Add(kb, ub)
	}
	vlog.Printf("defRefUnitsIndex: building phtable index...")
	h, err := b.Build()
	if err != nil {
		return err
	}
	x.phtable = h
	x.ready = true
	vlog.Printf("defRefUnitsIndex: done building index.")
	return nil
}

// Write implements persistedIndex.
func (x *defRefUnitsIndex) Write(w io.Writer) error {
	x.RLock()
	defer x.RUnlock()
	if x.phtable == nil {
		panic("no phtable to write")
	}
	return x.phtable.Write(w)
}

// Read implements persistedIndex.
func (x *defRefUnitsIndex) Read(r io.Reader) error {
	phtable, err := phtable.Read(r)
	x.Lock()
	defer x.Unlock()
	x.phtable = phtable
	x.ready = (err == nil)
	return err
}

// Ready implements persistedIndex.
func (x *defRefUnitsIndex) Ready() bool {
	x.RLock()
	defer x.RUnlock()
	return x.ready
}
