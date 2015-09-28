package store

import (
	"fmt"
	"io"
	"path"

	"github.com/alecthomas/binary"

	"sourcegraph.com/sourcegraph/srclib/store/phtable"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// NOTE(sqs): There is a lot of duplication here with defFilesIndex.

// unitFilesIndex makes it fast to determine which source units
// contain a file (or files in a dir).
type unitFilesIndex struct {
	phtable *phtable.CHD
	ready   bool
}

var _ interface {
	Index
	persistedIndex
	unitIndexBuilder
	unitIndex
} = (*unitFilesIndex)(nil)

var c_unitFilesIndex_getByPath = 0 // counter

func (x *unitFilesIndex) String() string { return fmt.Sprintf("unitFilesIndex(ready=%v)", x.ready) }

// getByFile returns a list of source units that contain the file
// specified by the path. The path can also be a directory, in which
// case all source units that contain files underneath that directory
// are returned.
func (x *unitFilesIndex) getByPath(path string) ([]unit.ID2, bool, error) {
	vlog.Printf("unitFilesIndex.getByPath(%s)", path)
	c_unitFilesIndex_getByPath++

	if x.phtable == nil {
		panic("phtable not built/read")
	}
	v := x.phtable.Get([]byte(path))
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
func (x *unitFilesIndex) Covers(filters interface{}) int {
	cov := 0
	for _, f := range storeFilters(filters) {
		if _, ok := f.(ByFilesFilter); ok {
			cov++
		}
	}
	return cov
}

// Units implements unitIndex.
func (x *unitFilesIndex) Units(fs ...UnitFilter) ([]unit.ID2, error) {
	for _, f := range fs {
		if ff, ok := f.(ByFilesFilter); ok {
			files := ff.ByFiles()
			umap := map[unit.ID2]struct{}{}
			for _, file := range files {
				u, _, err := x.getByPath(file)
				if err != nil {
					return nil, err
				}
				for _, uu := range u {
					umap[uu] = struct{}{}
				}
			}
			us := make([]unit.ID2, 0, len(umap))
			for u := range umap {
				us = append(us, u)
			}

			vlog.Printf("unitFilesIndex(%v): Found units %v using index.", fs, us)
			return us, nil
		}
	}
	return nil, nil
}

// Build implements unitIndexBuilder.
func (x *unitFilesIndex) Build(units []*unit.SourceUnit) error {
	vlog.Printf("unitFilesIndex: building index...")
	f2u := make(filesToUnits, len(units)*10)
	for _, u := range units {
		for _, f := range u.Files {
			f2u.add(f, u.ID2())
		}
	}
	b := phtable.Builder(len(f2u))
	for file, fileUnits := range f2u {
		ub, err := binary.Marshal(fileUnits)
		if err != nil {
			return err
		}
		b.Add([]byte(file), ub)
	}
	h, err := b.Build()
	if err != nil {
		return err
	}
	x.phtable = h
	x.ready = true
	vlog.Printf("unitFilesIndex: done building index.")
	return nil
}

// filesToUnits is a helper type used by unitFilesIndex.Build that
// adds parent dirs of each file to the mapping as well.
//
// TODO(sqs): lots of duplication with filesToDefOfs
type filesToUnits map[string][]unit.ID2

// add appends u to file's list of units, as well as the list of units
// for each of file's ancestor dirs.
func (v filesToUnits) add(file string, u unit.ID2) {
	file = path.Clean(file)
	v[file] = append(v[file], u)
	for _, dir := range ancestorDirsExceptRoot(file) {
		v.addIfNotExists(dir, u)
	}
}

// addIfNotExists appends u to dir's list of units, if u is not
// already present in the list.
func (v filesToUnits) addIfNotExists(dir string, u unit.ID2) {
	for _, uu := range v[dir] {
		if u == uu {
			return
		}
	}
	v[dir] = append(v[dir], u)
}

// ancestorDirsExceptRoot returns a list of p's ancestor directories
// excluding the root ("." or "/").
func ancestorDirsExceptRoot(p string) []string {
	if p == "" {
		return nil
	}
	if len(p) == 1 && (p[0] == '.' || p[0] == '/') {
		return nil
	}

	var dirs []string
	for i, c := range p {
		if c == '/' {
			dirs = append(dirs, p[:i])
		}
	}
	return dirs
}

// Write implements persistedIndex.
func (x *unitFilesIndex) Write(w io.Writer) error {
	if x.phtable == nil {
		panic("no phtable to write")
	}
	return x.phtable.Write(w)
}

// Read implements persistedIndex.
func (x *unitFilesIndex) Read(r io.Reader) error {
	var err error
	x.phtable, err = phtable.Read(r)
	x.ready = (err == nil)
	return err
}

// Ready implements persistedIndex.
func (x *unitFilesIndex) Ready() bool { return x.ready }
