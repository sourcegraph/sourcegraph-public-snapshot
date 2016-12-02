package store

import (
	"io"

	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/store/phtable"
)

type defPathIndex struct {
	phtable *phtable.CHD
	ready   bool
}

var _ interface {
	Index
	persistedIndex
	defIndexBuilder
	defIndex
} = (*defPathIndex)(nil)

func (x *defPathIndex) getByPath(defPath string) (int64, bool) {
	if x.phtable == nil {
		panic("phtable not built/read")
	}
	v, found := x.phtable.GetUint64([]byte(defPath))
	return int64(v), found
}

// Covers implements defIndex.
func (x *defPathIndex) Covers(filters interface{}) int {
	cov := 0
	for _, f := range storeFilters(filters) {
		if _, ok := f.(ByDefPathFilter); ok {
			cov++
		}
	}
	return cov
}

// Defs implements defIndex.
func (x *defPathIndex) Defs(f ...DefFilter) (byteOffsets, error) {
	for _, ff := range f {
		if pf, ok := ff.(ByDefPathFilter); ok {
			ofs, found := x.getByPath(pf.ByDefPath())
			if !found {
				return nil, nil
			}
			return byteOffsets{ofs}, nil
		}
	}
	return nil, nil
}

// Build implements defIndexBuilder.
func (x *defPathIndex) Build(defs []*graph.Def, ofs byteOffsets) error {
	tries := 0
retry:
	vlog.Printf("defPathIndex: building index... (%d defs)", len(defs))
	b := phtable.Uvarint64Builder(len(defs))
	for i, def := range defs {
		b.AddUvarint64([]byte(def.Path), uint64(ofs[i]))
	}
	vlog.Printf("defPathIndex: done adding index (%d defs).", len(defs))
	h, err := b.Build()
	if err != nil {
		if tries < 10 && strings.Contains(err.Error(), "failed to find a collision-free hash function") {
			tries++
			goto retry
		}
		return err
	}
	h.ValuesAreVarints = true
	x.phtable = h
	x.ready = true
	vlog.Printf("defPathIndex: done building index (%d defs).", len(defs))
	return nil
}

// Write implements persistedIndex.
func (x *defPathIndex) Write(w io.Writer) error {
	if x.phtable == nil {
		panic("no phtable to write")
	}
	return x.phtable.Write(w)
}

// Read implements persistedIndex.
func (x *defPathIndex) Read(r io.Reader) error {
	var err error
	x.phtable, err = phtable.ReadVarints(r)
	x.ready = (err == nil)
	return err
}

// Ready implements persistedIndex.
func (x *defPathIndex) Ready() bool { return x.ready }
