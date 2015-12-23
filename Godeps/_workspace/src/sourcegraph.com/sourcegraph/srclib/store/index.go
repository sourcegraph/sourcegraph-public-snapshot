package store

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// An Index enables efficient store queries using filters that the
// index covers. An index may be in one of 3 states:
//
//  * Not built: the index neither exists in memory nor is it
//    persisted. It can't be used.
//
//  * Persisted but not ready: the index has been built and persisted
//    (e.g., to disk) but has not been loaded into memory and therefore
//    can't be used.
//
//  * Ready: the index is loaded into memory (either because it was
//    just built in memory, or because it was read from its persisted
//    form) and can be used.
type Index interface {
	// Ready indicates whether the index is ready to be
	// queried. Persisted indexes typically become ready after their
	// Read method is called and returns.
	Ready() bool

	// Covers returns the number of filters that this index
	// covers. Indexes with greater coverage are selected over others
	// with lesser coverage.
	Covers(filters interface{}) int
}

// A persistedIndex is an index that can be serialized and
// deserialized.
type persistedIndex interface {
	// Write serializes an index to a writer. The index's Read method
	// can be called to deserialize the index at a later date.
	Write(io.Writer) error

	// Read populates an index from a reader that contains the same
	// data that the index previously wrote (using Write).
	Read(io.Reader) error
}

// The rest of this file contains helpers used by many index
// implementations.

type byteOffsets []int64

func (v byteOffsets) MarshalBinary() ([]byte, error) {
	bb := make([]byte, len(v)*binary.MaxVarintLen64)
	b := bb
	for _, ofs := range v {
		n := binary.PutVarint(b, ofs)
		b = b[n:]
	}
	return bb[:len(bb)-len(b)], nil
}

func (v *byteOffsets) UnmarshalBinary(b []byte) error {
	for {
		if len(b) == 0 {
			break
		}
		ofs, n := binary.Varint(b)
		if n == 0 {
			return io.ErrShortBuffer
		}
		if n < 0 {
			return errors.New("bad varint")
		}
		*v = append(*v, ofs)
		b = b[n:]
	}
	return nil
}

type defIndexBuilder interface {
	Build([]*graph.Def, byteOffsets) error
}

type defIndex interface {
	// Defs returns the byte offsets (within the def data file) of the
	// defs that match the def filters.
	Defs(...DefFilter) (byteOffsets, error)
}

type defTreeIndex interface {
	// Defs returns the source units and byte offsets (within the
	// source unit def data file) of the defs that match the def
	// filters.
	Defs(...DefFilter) (map[unit.ID2]byteOffsets, error)
}

// bestCoverageIndex returns the index that has the greatest coverage
// for the given filters, or nil if no indexes have any coverage. If
// test != nil, only indexes for which test(x) is true are considered.
func bestCoverageIndex(indexes map[string]Index, filters interface{}, test func(x interface{}) bool) (bestName string, best Index) {
	bestCov := 0
	for name, x := range indexes {
		if test != nil && !test(x) {
			continue
		}
		cov := x.Covers(filters)
		if cov > bestCov {
			bestCov = cov
			bestName = name
			best = x
		}
	}
	return bestName, best
}

func isUnitIndex(x interface{}) bool    { _, ok := x.(unitIndex); return ok }
func isDefIndex(x interface{}) bool     { _, ok := x.(defIndex); return ok }
func isDefTreeIndex(x interface{}) bool { _, ok := x.(defTreeIndex); return ok }
func isRefIndex(x interface{}) bool {
	switch x.(type) {
	case refIndexByteRanges, refIndexByteOffsets:
		return true
	}
	return false
}

// fileByteRanges maps from filename to the byte ranges in a byte
// array that pertain to that file. It's used to index into the ref
// data to quickly read all of the refs in a given file.
type fileByteRanges map[string]byteRanges

// byteRanges' encodes the byte offsets of multiple objects. The first
// element is the byte offset within a file. Subsequent elements are
// the byte length of each object in the file.
type byteRanges []int64

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (b *byteRanges) UnmarshalBinary(data []byte) error {
	for {
		v, n := binary.Varint(data)
		if n == 0 {
			break
		}
		if n < 0 {
			return errors.New("byteRanges varint error")
		}
		*b = append(*b, v)
		data = data[n:]
	}
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (b byteRanges) MarshalBinary() ([]byte, error) {
	data := make([]byte, len(b)*binary.MaxVarintLen64)
	var n int
	for _, v := range b {
		n += binary.PutVarint(data[n:], v)
	}
	return data[:n], nil
}

// start is the offset of the first byte of the first object, relative
// to the beginning of the file.
func (br byteRanges) start() int64 { return br[0] }

// refsByFileStartEnd sorts refs by (file, start, end).
type refsByFileStartEnd []*graph.Ref

func (v refsByFileStartEnd) Len() int { return len(v) }
func (v refsByFileStartEnd) Less(i, j int) bool {
	a, b := v[i], v[j]
	return a.File < b.File || (a.File == b.File && a.Start < b.Start) || (a.File == b.File && a.Start == b.Start && a.End < b.End)
}
func (v refsByFileStartEnd) Less2(i, j int) bool {
	a, b := v[i], v[j]
	if a.File == b.File {
		if a.Start == b.Start {
			return a.End < b.End
		}
		return a.Start < b.Start
	}
	return a.File < b.File
}
func (v refsByFileStartEnd) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

type refIndexByteRanges interface {
	// Refs returns the byte ranges (in the ref data file) of matching
	// refs, or errNotIndexed if no indexes can be used to satisfy the
	// query.
	Refs(...RefFilter) ([]byteRanges, error)
}

type refIndexByteOffsets interface {
	// Refs returns the byte offsets (in the ref data file) of matching
	// refs, or errNotIndexed if no indexes can be used to satisfy the
	// query.
	Refs(...RefFilter) (byteOffsets, error)
}

type refIndexBuilder interface {
	// Build constructs the index in memory.
	Build([]*graph.Ref, fileByteRanges, byteOffsets) error
}

type unitIndex interface {
	// Units returns the unit IDs units that match the unit filters.
	Units(...UnitFilter) ([]unit.ID2, error)
}

type unitFullIndex interface {
	Units(...UnitFilter) ([]*unit.SourceUnit, error)
}

type unitIndexBuilder interface {
	// Build constructs the index in memory.
	Build([]*unit.SourceUnit) error
}

type unitRefIndexBuilder interface {
	Build(map[unit.ID2]*defRefsIndex) error
}

type defQueryTreeIndexBuilder interface {
	Build(map[unit.ID2]*defQueryIndex) error
}

// unitIndexOnlyFilter wraps a non-UnitFilter that can be used by an
// IndexedUnitStore to scope the list of source units. Currently there
// is only a RefFilter that does this, so we simplify it by using that
// concrete type as the field type.
type unitIndexOnlyFilter struct{ ByRefDefFilter }

func (f unitIndexOnlyFilter) SelectUnit(u *unit.SourceUnit) bool {
	// Index-only filter; can't determine selection with information
	// available to filter. So assume that if this filter is being
	// used, the index has already scoped the results and u was
	// selected.
	return true
}

// unitOffsets holds a set of byte offsets that all refer to positions
// in a file inside a specific source unit.
type unitOffsets struct {
	byteOffsets        // byte offsets of defs/refs/etc. in source unit data file (def.dat, ref.dat, etc.)
	Unit        uint16 // index of source unit
}

func (v *unitOffsets) MarshalBinary() ([]byte, error) {
	ofsB, err := v.byteOffsets.MarshalBinary()
	if err != nil {
		return nil, err
	}
	ofsB = append(ofsB, byte(0), byte(0))
	binary.LittleEndian.PutUint16(ofsB[len(ofsB)-2:], v.Unit)

	return ofsB, nil
}

func (v *unitOffsets) UnmarshalBinary(b []byte) error {
	v.Unit = binary.LittleEndian.Uint16(b[len(b)-2:])
	return v.byteOffsets.UnmarshalBinary(b[:len(b)-2])
}

// unitDefOffsetsFilter is an internal filter used by indexes. It
// selects only defs at certain byte offsets in certain source units.
type unitDefOffsetsFilter map[unit.ID2]byteOffsets

var _ interface {
	ByUnitsFilter
	UnitFilter
	DefFilter
} = (*unitDefOffsetsFilter)(nil)

func (f unitDefOffsetsFilter) String() string {
	return fmt.Sprintf("unitDefOffsetsFilter(%v)", map[unit.ID2]byteOffsets(f))
}

func (f unitDefOffsetsFilter) ByUnits() []unit.ID2 {
	units := make([]unit.ID2, 0, len(f))
	for u := range f {
		units = append(units, u)
	}
	return units
}

func (f unitDefOffsetsFilter) SelectUnit(u *unit.SourceUnit) bool {
	_, present := f[u.ID2()]
	return present
}

func (f unitDefOffsetsFilter) SelectDef(*graph.Def) bool {
	// Index-only filter; can't determine selection with information
	// available to filter. So assume that if this filter is being
	// used, the index has already scoped the results to defs that it
	// would select.
	return true
}

// defOffsetsFilter is an internal filter used by indexes. It
// selects only defs at certain byte offsets in the def.dat file.
type defOffsetsFilter byteOffsets

var _ interface {
	DefFilter
} = (*defOffsetsFilter)(nil)

func (f defOffsetsFilter) String() string {
	return fmt.Sprintf("defOffsetsFilter(%v)", byteOffsets(f))
}

func (f defOffsetsFilter) SelectDef(*graph.Def) bool {
	// Index-only filter; can't determine selection with information
	// available to filter. So assume that if this filter is being
	// used, the index has already scoped the results to defs that it
	// would select.
	return true
}

// getDefOffsetsFilter returns a defOffsetsFilter in fs, if any exists. Otherwise it returns nil.
func getDefOffsetsFilter(fs []DefFilter) defOffsetsFilter {
	for _, f := range fs {
		if f, ok := f.(defOffsetsFilter); ok {
			return f
		}
	}
	return nil
}
