package store

import (
	"io"

	"github.com/alecthomas/binary"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/store/phtable"
)

// refFileIndex makes it fast to determine which refs (within in a
// source unit) are in a file.
type refFileIndex struct {
	phtable *phtable.CHD
	ready   bool
}

var _ interface {
	Index
	persistedIndex
	refIndexByteRanges
	refIndexBuilder
} = (*refFileIndex)(nil)

var c_refFileIndex_getByFile = 0 // counter

// getByFile returns a byteRanges describing the positions of refs in
// the given source file (i.e., for which ref.File == file). The
// byteRanges refer to offsets within the ref data file.
func (x *refFileIndex) getByFile(file string) (byteRanges, bool, error) {
	c_refFileIndex_getByFile++
	if x.phtable == nil {
		panic("phtable not built/read")
	}
	v := x.phtable.Get([]byte(file))
	if v == nil {
		return nil, false, nil
	}

	var br byteRanges
	if err := binary.Unmarshal(v, &br); err != nil {
		return nil, true, err
	}
	return br, true, nil
}

// Covers implements defIndex.
func (x *refFileIndex) Covers(filters interface{}) int {
	// TODO(sqs): this index also covers RefStart/End range filters
	// (when those are added).
	cov := 0
	for _, f := range storeFilters(filters) {
		if _, ok := f.(ByFilesFilter); ok {
			cov++
		}
	}
	return cov
}

// Refs implements refIndexByteRanges.
func (x *refFileIndex) Refs(fs ...RefFilter) ([]byteRanges, error) {
	for _, f := range fs {
		if ff, ok := f.(ByFilesFilter); ok {
			files := ff.ByFiles()
			brs := make([]byteRanges, 0, len(files))
			for _, file := range files {
				br, found, err := x.getByFile(file)
				if err != nil {
					return nil, err
				}
				if found {
					brs = append(brs, br)
				}
			}
			return brs, nil
		}
	}
	return nil, nil
}

// Build creates the refFileIndex.
func (x *refFileIndex) Build(_ []*graph.Ref, fbr fileByteRanges, _ byteOffsets) error {
	vlog.Printf("refFilesIndex: building index...")
	b := phtable.Builder(len(fbr))
	for file, br := range fbr {
		v, err := binary.Marshal(br)
		if err != nil {
			return err
		}
		b.Add([]byte(file), v)
	}
	h, err := b.Build()
	if err != nil {
		return err
	}
	x.phtable = h
	x.ready = true
	vlog.Printf("refFilesIndex: done building index.")
	return nil
}

// Write implements persistedIndex.
func (x *refFileIndex) Write(w io.Writer) error {
	if x.phtable == nil {
		panic("no phtable to write")
	}
	return x.phtable.Write(w)
}

// Read implements persistedIndex.
func (x *refFileIndex) Read(r io.Reader) error {
	var err error
	x.phtable, err = phtable.Read(r)
	x.ready = (err == nil)
	return err
}

// Ready implements persistedIndex.
func (x *refFileIndex) Ready() bool { return x.ready }
