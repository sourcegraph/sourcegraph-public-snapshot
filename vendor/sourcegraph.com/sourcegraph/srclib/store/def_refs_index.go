package store

import (
	"io"
	"sync"

	"github.com/alecthomas/binary"
	"github.com/gogo/protobuf/proto"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/store/phtable"
)

// defRefsIndex makes it fast to determine which refs (within in a
// source unit) are in a file.
type defRefsIndex struct {
	phtable *phtable.CHD
	ready   bool
	sync.RWMutex
}

var _ interface {
	Index
	persistedIndex
	refIndexByteOffsets
	refIndexBuilder
} = (*defRefsIndex)(nil)

var c_defRefsIndex_getByDef = &counter{count: new(int64)}

func (x *defRefsIndex) String() string { return "defRefsIndex" }

func (x *defRefsIndex) getByDef(def graph.RefDefKey) (byteOffsets, bool, error) {
	c_defRefsIndex_getByDef.increment()
	if x.phtable == nil {
		panic("phtable not built/read")
	}

	k, err := proto.Marshal(&def)
	if err != nil {
		return nil, false, err
	}

	v := x.phtable.Get(k)
	if v == nil {
		return nil, false, nil
	}

	var ofs byteOffsets
	if err := binary.Unmarshal(v, &ofs); err != nil {
		return nil, true, err
	}
	return ofs, true, nil
}

// Covers implements refIndex.
func (x *defRefsIndex) Covers(filters interface{}) int {
	cov := 0
	for _, f := range storeFilters(filters) {
		if _, ok := f.(ByRefDefFilter); ok {
			cov++
		}
	}
	return cov
}

// Refs implements refIndexByteOffsets.
func (x *defRefsIndex) Refs(fs ...RefFilter) (byteOffsets, error) {
	x.RLock()
	defer x.RUnlock()
	for _, f := range fs {
		if ff, ok := f.(ByRefDefFilter); ok {
			ofs, found, err := x.getByDef(ff.withEmptyImpliedValues())
			if err != nil {
				return nil, err
			}
			if found {
				return ofs, nil
			}
		}
	}
	return nil, nil
}

// Build creates the defRefsIndex.
func (x *defRefsIndex) Build(refs []*graph.Ref, fbr fileByteRanges, ofs byteOffsets) error {
	x.Lock()
	defer x.Unlock()
	vlog.Printf("defRefsIndex: building inverted def->ref index (%d refs)...", len(refs))
	defToRefOfs := map[graph.RefDefKey]byteOffsets{}
	for i, ref := range refs {
		defToRefOfs[ref.RefDefKey()] = append(defToRefOfs[ref.RefDefKey()], ofs[i])
	}

	vlog.Printf("defRefsIndex: adding %d index phtable keys...", len(defToRefOfs))
	b := phtable.Builder(len(fbr))
	for def, refOfs := range defToRefOfs {
		v, err := binary.Marshal(refOfs)
		if err != nil {
			return err
		}

		k, err := proto.Marshal(&def)
		if err != nil {
			return err
		}

		b.Add([]byte(k), v)
	}
	vlog.Printf("defRefsIndex: building index phtable...")
	h, err := b.Build()
	if err != nil {
		return err
	}
	h.StoreKeys = true // so defRefUnitsIndex can enumerate defs pointed to by this unit's refs
	x.phtable = h
	x.ready = true
	vlog.Printf("defRefsIndex: done building index.")
	return nil
}

// Write implements persistedIndex.
func (x *defRefsIndex) Write(w io.Writer) error {
	x.RLock()
	defer x.RUnlock()
	if x.phtable == nil {
		panic("no phtable to write")
	}
	return x.phtable.Write(w)
}

// Read implements persistedIndex.
func (x *defRefsIndex) Read(r io.Reader) error {
	phtable, err := phtable.Read(r)
	x.Lock()
	defer x.Unlock()
	x.phtable = phtable
	x.ready = (err == nil)
	return err
}

// Ready implements persistedIndex.
func (x *defRefsIndex) Ready() bool {
	x.RLock()
	defer x.RUnlock()
	return x.ready
}
