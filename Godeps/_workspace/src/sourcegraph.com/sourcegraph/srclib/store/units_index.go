package store

import (
	"encoding/json"
	"fmt"
	"io"

	"sourcegraph.com/sourcegraph/srclib/unit"
)

// unitsIndex makes it fast list all source units without needing to
// traverse directories.
type unitsIndex struct {
	units []*unit.SourceUnit
	ready bool
}

var _ interface {
	Index
	persistedIndex
	unitIndexBuilder
	unitFullIndex
} = (*unitsIndex)(nil)

var c_unitsIndex_listUnits = 0 // counter

func (x *unitsIndex) String() string { return fmt.Sprintf("unitsIndex(ready=%v)", x.ready) }

// Covers returns -1 because it should never be selected over a more
// specific index. When the indexed stores use the unitsIndex, they do
// so through a different code path from the one that selects other
// indexes.
func (x *unitsIndex) Covers(filters interface{}) int { return -1 }

// Units implements unitFullIndex.
func (x *unitsIndex) Units(fs ...UnitFilter) ([]*unit.SourceUnit, error) {
	if x.units == nil {
		panic("units not built/read")
	}

	c_unitsIndex_listUnits++

	var units []*unit.SourceUnit
	for _, u := range x.units {
		if unitFilters(fs).SelectUnit(u) {
			units = append(units, u)
		}
	}

	return units, nil
}

// Build implements unitIndexBuilder.
func (x *unitsIndex) Build(units []*unit.SourceUnit) error {
	x.units = units
	x.ready = true
	return nil
}

// Write implements persistedIndex.
func (x *unitsIndex) Write(w io.Writer) error {
	if x.units == nil {
		panic("no units to write")
	}
	return json.NewEncoder(w).Encode(x.units)
}

// Read implements persistedIndex.
func (x *unitsIndex) Read(r io.Reader) error {
	err := json.NewDecoder(r).Decode(&x.units)
	x.ready = (err == nil)
	return err
}

// Ready implements persistedIndex.
func (x *unitsIndex) Ready() bool { return x.ready }
