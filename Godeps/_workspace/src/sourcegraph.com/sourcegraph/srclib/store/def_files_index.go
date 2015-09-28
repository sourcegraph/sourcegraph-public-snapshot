package store

import (
	"fmt"
	"io"
	"path"
	"sync"

	"github.com/alecthomas/binary"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/store/phtable"
)

// NOTE(sqs): There is a lot of duplication here with unitFilesIndex.

// defFilesIndex makes it fast to determine which source units
// contain a file (or files in a dir).
type defFilesIndex struct {
	// filters restricts the index to only indexing defs that pass all
	// of the filters.
	filters []DefFilter

	// perFile is the number of defs per file to index.
	perFile int

	phtable *phtable.CHD
	ready   bool

	sync.RWMutex
}

var _ interface {
	Index
	persistedIndex
	defIndexBuilder
	defIndex
} = (*defFilesIndex)(nil)

var c_defFilesIndex_getByPath = 0 // counter

func (x *defFilesIndex) String() string {
	return fmt.Sprintf("defFilesIndex(ready=%v, filters=%v)", x.ready, x.filters)
}

// getByFile returns a list of source units that contain the file
// specified by the path. The path can also be a directory, in which
// case all source units that contain files underneath that directory
// are returned.
func (x *defFilesIndex) getByPath(path string) (byteOffsets, bool, error) {
	vlog.Printf("defFilesIndex.getByPath(%s)", path)
	c_defFilesIndex_getByPath++

	if x.phtable == nil {
		panic("phtable not built/read")
	}
	v := x.phtable.Get([]byte(path))
	if v == nil {
		return nil, false, nil
	}

	var ofs byteOffsets
	if err := binary.Unmarshal(v, &ofs); err != nil {
		return nil, true, err
	}
	return ofs, true, nil
}

// Covers implements defIndex.
func (x *defFilesIndex) Covers(filters interface{}) int {
	// TODO(sqs): ensure that x.filters is equivalent to fs (might
	// require an equals() method on filters, for filters with
	// internal state that we don't necessarily want to use when
	// testing equality). Otherwise this just assumes that fs has a
	// Limit, an Exported=true/Nonlocal=true filter, etc.
	cov := 0
	for _, f := range storeFilters(filters) {
		if _, ok := f.(ByFilesFilter); ok {
			cov++
		}
	}
	return cov
}

// Defs implements defIndex.
func (x *defFilesIndex) Defs(fs ...DefFilter) (byteOffsets, error) {
	x.RLock()
	defer x.RUnlock()
	for _, f := range fs {
		if ff, ok := f.(ByFilesFilter); ok {
			files := ff.ByFiles()
			var allOfs byteOffsets
			for _, file := range files {
				ofs, _, err := x.getByPath(file)
				if err != nil {
					return nil, err
				}
				allOfs = append(allOfs, ofs...)
			}

			vlog.Printf("defFilesIndex(%v): Found %d def offsets using index.", fs, len(allOfs))
			return allOfs, nil
		}
	}
	return nil, nil
}

// Build implements defIndexBuilder.
func (x *defFilesIndex) Build(defs []*graph.Def, ofs byteOffsets) error {
	x.Lock()
	defer x.Unlock()
	vlog.Printf("defFilesIndex: building index...")
	f2ofs := make(filesToDefOfs, len(defs)/50)
	for i, def := range defs {
		if len(f2ofs[def.File]) < x.perFile && DefFilters(x.filters).SelectDef(def) {
			f2ofs.add(def.File, ofs[i], x.perFile)
		}
	}
	b := phtable.Builder(len(f2ofs))
	for file, defOfs := range f2ofs {
		ob, err := binary.Marshal(defOfs)
		if err != nil {
			return err
		}
		b.Add([]byte(file), ob)
	}
	h, err := b.Build()
	if err != nil {
		return err
	}
	x.phtable = h
	x.ready = true
	vlog.Printf("defFilesIndex: done building index.")
	return nil
}

// filesToDefOfs is a helper type used by defFilesIndex.Build that
// adds parent dirs of each file to the mapping as well.
//
// TODO(sqs): lots of duplication with filesToUnits
type filesToDefOfs map[string]byteOffsets

// add appends ofs to file's list of def offsets, as well as the list
// of def offsets for each of file's ancestor dirs. If an entry has
// more than perFile offsets already, no more are appended.
func (v filesToDefOfs) add(file string, ofs int64, perFile int) {
	file = path.Clean(file)
	if len(v[file]) >= perFile {
		return
	}
	v[file] = append(v[file], ofs)
	for _, dir := range ancestorDirsExceptRoot(file) {
		if len(v[dir]) < perFile {
			v[dir] = append(v[dir], ofs)
		}
	}
}

// Write implements persistedIndex.
func (x *defFilesIndex) Write(w io.Writer) error {
	x.RLock()
	defer x.RUnlock()
	if x.phtable == nil {
		panic("no phtable to write")
	}
	return x.phtable.Write(w)
}

// Read implements persistedIndex.
func (x *defFilesIndex) Read(r io.Reader) error {
	phtable, err := phtable.Read(r)
	x.Lock()
	defer x.Unlock()
	x.phtable = phtable
	x.ready = (err == nil)
	return err
}

// Ready implements persistedIndex.
func (x *defFilesIndex) Ready() bool {
	x.RLock()
	defer x.RUnlock()
	return x.ready
}
