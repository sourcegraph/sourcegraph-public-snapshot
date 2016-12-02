package store

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"

	"golang.org/x/tools/godoc/vfs"

	"compress/gzip"

	"github.com/neelance/parallel"

	"sort"
	"strings"
	"sync"

	"sourcegraph.com/sourcegraph/rwvfs"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

type indexedStore interface {
	// Indexes lists all indexes that have been built.
	Indexes() map[string]Index

	// BuildIndex builds the index with the specified name.
	BuildIndex(name string, x Index) error

	// readIndex calls the readIndex func on the given index.
	readIndex(name string, x persistedIndex) error

	// statIndex calls vfs.Stat on the index's backing file or
	// directory.
	statIndex(name string) (os.FileInfo, error)
}

// An indexedTreeStore is a VFS-backed tree store that generates
// indexes to provide efficient lookups.
//
// It wraps a fsTreeStore and intercepts calls to Def, Defs,
// Refs, etc., using its indexes to satisfy those queries efficiently
// where possible. Otherwise it invokes the underlying store to
// perform a full scan (over all source units in the tree).
type indexedTreeStore struct {
	// indexes is all of the indexes that should be built, written,
	// and read from. It contains all indexes for all types of data
	// (e.g., def indexes, ref indexes, etc.).
	indexes map[string]Index

	// cacheKey uniquely identifies the collection of indexes. This is
	// used to sharing indexes in concurrent queries
	cacheKey interface{}

	*fsTreeStore
}

var _ interface {
	TreeStore
	indexedStore
} = (*indexedTreeStore)(nil)

const (
	unitsIndexName = "units"
)

// newIndexedTreeStore creates a new indexed tree store that stores
// data and indexes in fs.
func newIndexedTreeStore(fs rwvfs.FileSystem, cacheKey interface{}) TreeStoreImporter {
	return &indexedTreeStore{
		indexes: map[string]Index{
			"file_to_units":       &unitFilesIndex{},
			"def_to_ref_units":    &defRefUnitsIndex{},
			"def_query_to_defs16": &defQueryTreeIndex{},
			unitsIndexName:        &unitsIndex{},
		},
		cacheKey:    cacheKey,
		fsTreeStore: newFSTreeStore(fs),
	}
}

func (s *indexedTreeStore) StoreKey() interface{} { return s.cacheKey }
func (s *indexedTreeStore) String() string        { return "indexedTreeStore" }

// errNotIndexed occurs when that a query was unable to be performed
// using an index. In most cases, it indicates that the caller should
// perform the query using a full scan of all of the data.
var errNotIndexed = errors.New("no index satisfies query")

// unitIDs returns the source unit IDs that satisfy the unit
// filters. If possible, it uses indexes instead of performing a full
// scan of the source unit files.
//
// If indexOnly is specified, only the index will be consulted. If a
// full scan would otherwise occur, errNotIndexed is returned.
func (s *indexedTreeStore) unitIDs(indexOnly bool, fs ...UnitFilter) ([]unit.ID2, error) {
	vlog.Printf("indexedTreeStore.unitIDs(indexOnly=%v, %v)", indexOnly, fs)

	scopedUnits, err := scopeUnits(storeFilters(fs))
	if err != nil {
		return nil, err
	}
	if scopedUnits != nil {
		vlog.Printf("indexedTreeStore.unitIDs(indexOnly=%v, %v): Returning scoped units (from filters) without performing external lookup.", indexOnly, fs)
		return scopedUnits, nil
	}

	// Try to find an index that covers this query.
	if xname, bx := bestCoverageIndex(s.indexes, fs, isUnitIndex); bx != nil {
		if !bx.Ready() {
			bx = cacheGet(s, xname, bx)
		}
		if err := prepareIndex(s.fs, xname, bx); err != nil {
			return nil, err
		}
		cachePut(s, xname, bx)
		vlog.Printf("indexedTreeStore.unitIDs(%v): Found covering index %q (%v).", fs, xname, bx)
		return bx.(unitIndex).Units(fs...)
	}
	if indexOnly {
		return nil, errNotIndexed
	}

	// Fall back to full scan.
	vlog.Printf("indexedTreeStore.unitIDs(%v): No covering indexes found; performing full scan.", fs)
	var unitIDs []unit.ID2
	units, err := s.unitsUsingFullIndex(fs...)
	if err != nil {
		return nil, err
	}
	for _, u := range units {
		unitIDs = append(unitIDs, u.ID2())
	}
	return unitIDs, nil
}

// maxIndividualFetches is a heuristic: If we need to fetch more than
// this many source unit files, fetching the unitsIndex (which has all
// source unit definitions in one file) is probably faster.
var maxIndividualFetches = 5

func (s *indexedTreeStore) Units(fs ...UnitFilter) ([]*unit.SourceUnit, error) {
	// Attempt to use the index.
	scopedUnits, err := s.unitIDs(true, fs...)
	if err != nil && err != errNotIndexed {
		return nil, err
	}

	if len(scopedUnits) == 0 || len(scopedUnits) > maxIndividualFetches {
		vlog.Printf("indexedTreeStore.Units(%v): Using unitsIndex for query scoped to %d units.", fs, len(scopedUnits))
		return s.unitsUsingFullIndex(fs...)
	}

	vlog.Printf("indexedTreeStore.Units(%v): Delegating to fsTreeStore for query scoped to %d units.", fs, len(scopedUnits))
	return s.fsTreeStore.Units(fs...)
}

func (s *indexedTreeStore) unitsUsingFullIndex(fs ...UnitFilter) ([]*unit.SourceUnit, error) {
	x := s.indexes[unitsIndexName]
	if err := prepareIndex(s.fs, unitsIndexName, x); err != nil {
		return nil, err
	}
	return x.(unitFullIndex).Units(fs...)
}

func (s *indexedTreeStore) Defs(fs ...DefFilter) ([]*graph.Def, error) {
	vlog.Printf("indexedTreeStore.Defs(%v)", fs)

	// First, check if any defs indexes at the tree level cover this
	// query.
	if xname, bx := bestCoverageIndex(s.indexes, fs, isDefTreeIndex); bx != nil {
		if err := prepareIndex(s.fs, xname, bx); err != nil {
			return nil, err
		}
		vlog.Printf("indexedTreeStore.Defs(%v): Found covering index %q (%v).", fs, xname, bx)
		uoffs, err := bx.(defTreeIndex).Defs(fs...)
		if err != nil {
			return nil, err
		}
		fs = append(fs, unitDefOffsetsFilter(uoffs))
	}

	// We have File->Unit index (that tells us which source units
	// include a given file). If there's a ByFiles DefFilter, then we
	// can convert that filter into a ByUnits scope filter (which is
	// more efficient) by consulting the File->Unit index.

	var ufs []UnitFilter
	for _, f := range fs {
		switch f := f.(type) {
		case UnitFilter:
			ufs = append(ufs, f)
		}
	}

	// No indexes found that we can exploit here; forward to the
	// underlying store.
	if len(ufs) == 0 {
		vlog.Printf("indexedTreeStore.Defs(%v): No unit indexes found to narrow scope; forwarding to underlying store.", fs)
		return s.fsTreeStore.Defs(fs...)
	}

	// Find which source units match the unit filters; we'll restrict
	// our defs query to those source units.
	scopeUnits, err := s.unitIDs(false, ufs...)
	if err != nil && err != errNotIndexed {
		return nil, err
	} else if err == nil {
		// Add ByUnits filters that were implied by ByFiles (and other
		// UnitFilters).
		//
		// If scopeUnits is empty, the empty ByUnits filter will result in
		// the query matching nothing, which is the desired behavior.
		vlog.Printf("indexedTreeStore.Defs(%v): Adding equivalent ByUnits filters to scope to units %+v.", fs, scopeUnits)
		fs = append(fs, ByUnits(scopeUnits...))
	}

	// Pass the now more narrowly scoped query onto the underlying store.
	return s.fsTreeStore.Defs(fs...)
}

func (s *indexedTreeStore) Refs(fs ...RefFilter) ([]*graph.Ref, error) {
	// We have File->Unit index (that tells us which source units
	// include a given file). If there's a ByFiles RefFilter, then we
	// can convert that filter into a ByUnits scope filter (which is
	// more efficient) by consulting the File->Unit index.

	var ufs []UnitFilter
	for _, f := range fs {
		switch f := f.(type) {
		case UnitFilter:
			ufs = append(ufs, f)

		case ByRefDefFilter:
			// HACK: Special-case the defRefUnitsIndex. It doesn't fit cleanly
			// into our existing index selection scheme.
			ufs = append(ufs, unitIndexOnlyFilter{f})
		}
	}

	// No indexes found that we can exploit here; forward to the
	// underlying store.
	if len(ufs) == 0 {
		vlog.Printf("indexedTreeStore.Refs(%v): No unit indexes found to narrow scope; forwarding to underlying store.", fs)
		return s.fsTreeStore.Refs(fs...)
	}

	// Find which source units match the unit filters; we'll restrict
	// our refs query to those source units.
	scopeUnits, err := s.unitIDs(false, ufs...)
	if err != nil {
		return nil, err
	}

	// Add ByUnits filters that were implied by ByFiles (and other
	// UnitFilters).
	//
	// If scopeUnits is empty, the empty ByUnits filter will result in
	// the query matching nothing, which is the desired behavior.
	vlog.Printf("indexedTreeStore.Refs(%v): Adding equivalent ByUnits filters to scope to units %+v.", fs, scopeUnits)
	fs = append(fs, ByUnits(scopeUnits...))

	// Pass the now more narrowly scoped query onto the underlying store.
	return s.fsTreeStore.Refs(fs...)
}

func (s *indexedTreeStore) Import(u *unit.SourceUnit, data graph.Output) error {
	s.checkSourceUnitFiles(u, data)
	if err := s.fsTreeStore.Import(u, data); err != nil {
		return err
	}
	return nil
}

func (s *indexedTreeStore) Index() error {
	return s.buildIndexes(s.Indexes(), nil, nil, nil)
}

func (s *indexedTreeStore) Indexes() map[string]Index { return s.indexes }

// checkSourceUnitFiles warns if any files appear in graph data but
// are not in u.Files.
func (s *indexedTreeStore) checkSourceUnitFiles(u *unit.SourceUnit, data graph.Output) {
	if u == nil {
		return
	}

	graphFiles := make(map[string]struct{}, len(u.Files))
	for _, def := range data.Defs {
		graphFiles[def.File] = struct{}{}
	}
	for _, ref := range data.Refs {
		graphFiles[ref.File] = struct{}{}
	}
	for _, doc := range data.Docs {
		graphFiles[doc.File] = struct{}{}
	}
	for _, ann := range data.Anns {
		graphFiles[ann.File] = struct{}{}
	}
	delete(graphFiles, "")

	unitFiles := make(map[string]struct{}, len(u.Files))
	for _, f := range u.Files {
		unitFiles[f] = struct{}{}
	}
	if u.Dir != "" {
		unitFiles[u.Dir] = struct{}{}
	}

	var missingFiles []string
	for f := range graphFiles {
		if _, present := unitFiles[f]; !present {
			missingFiles = append(missingFiles, f)
		}
	}
	if len(missingFiles) > 0 {
		sort.Strings(missingFiles)
		log.Printf("Warning: The graph output (defs/refs/docs/anns) for source unit %+v contain %d references to files that are not present in the source unit's Files list. Indexed lookups by any of these missing files will return no results. To fix this, ensure that the source unit's Files list includes all files that appear in the graph output. The missing files are: %s.", u.ID2(), len(missingFiles), strings.Join(missingFiles, " "))
	}
}

func (s *indexedTreeStore) BuildIndex(name string, x Index) error {
	return s.buildIndexes(map[string]Index{name: x}, nil, nil, nil)
}

func (s *indexedTreeStore) readIndex(name string, x persistedIndex) error {
	return readIndex(s.fs, name, x)
}

func (s *indexedTreeStore) buildIndexes(xs map[string]Index, units []*unit.SourceUnit, unitRefIndexes map[unit.ID2]*defRefsIndex, unitDefQueryIndexes map[unit.ID2]*defQueryIndex) error {
	// TODO(sqs): there's a race condition here if multiple imports
	// are running concurrently, they could clobber each other's
	// indexes. (S3 is eventually consistent.)

	var getUnitsErr error
	var getUnitsOnce sync.Once
	getUnits := func() ([]*unit.SourceUnit, error) {
		getUnitsOnce.Do(func() {
			if getUnitsErr == nil && units == nil {
				units, getUnitsErr = s.fsTreeStore.Units()
			}
			if units == nil {
				units = []*unit.SourceUnit{}
			}
		})
		return units, getUnitsErr
	}

	var getUnitRefIndexesErr error
	var getUnitRefIndexesOnce sync.Once
	var unitRefIndexesLock sync.Mutex
	getUnitRefIndexes := func() (map[unit.ID2]*defRefsIndex, error) {
		getUnitRefIndexesOnce.Do(func() {
			if getUnitRefIndexesErr == nil && unitRefIndexes == nil {
				// Read in the defRefsIndex for all source units.
				units, err := getUnits()
				if err != nil {
					getUnitRefIndexesErr = err
					return
				}

				// Use openUnitStore on the list from getUnits so we
				// don't need to traverse the FS tree to enumerate all
				// the source units again (which is slow).
				uss := make(map[unit.ID2]UnitStore, len(units))
				for _, u := range units {
					uss[u.ID2()] = s.fsTreeStore.openUnitStore(u.ID2())
				}

				unitRefIndexes = make(map[unit.ID2]*defRefsIndex, len(units))
				par := parallel.NewRun(runtime.GOMAXPROCS(0))
				for u_, us_ := range uss {
					u := u_
					us, ok := us_.(*indexedUnitStore)
					if !ok {
						continue
					}

					par.Acquire()
					go func() {
						defer par.Release()
						x := us.indexes[defToRefsIndexName]
						if err := prepareIndex(us.fs, defToRefsIndexName, x); err != nil {
							par.Error(err)
							return
						}
						unitRefIndexesLock.Lock()
						defer unitRefIndexesLock.Unlock()
						unitRefIndexes[u] = x.(*defRefsIndex)
					}()
				}
				getUnitRefIndexesErr = par.Wait()
			}
			if unitRefIndexes == nil {
				unitRefIndexes = map[unit.ID2]*defRefsIndex{}
			}
		})
		return unitRefIndexes, getUnitRefIndexesErr
	}

	var getUnitDefQueryIndexesErr error
	var getUnitDefQueryIndexesOnce sync.Once
	var unitDefQueryIndexesLock sync.Mutex
	getUnitDefQueryIndexes := func() (map[unit.ID2]*defQueryIndex, error) {
		getUnitDefQueryIndexesOnce.Do(func() {
			if getUnitDefQueryIndexesErr == nil && unitDefQueryIndexes == nil {
				// Read in the defQueryIndex for all source units.
				units, err := getUnits()
				if err != nil {
					getUnitDefQueryIndexesErr = err
					return
				}

				// Use openUnitStore on the list from getUnits so we
				// don't need to traverse the FS tree to enumerate all
				// the source units again (which is slow).
				uss := make(map[unit.ID2]UnitStore, len(units))
				for _, u := range units {
					uss[u.ID2()] = s.fsTreeStore.openUnitStore(u.ID2())
				}

				unitDefQueryIndexes = make(map[unit.ID2]*defQueryIndex, len(units))
				par := parallel.NewRun(runtime.GOMAXPROCS(0))
				for u_, us_ := range uss {
					u := u_
					us, ok := us_.(*indexedUnitStore)
					if !ok {
						continue
					}

					par.Acquire()
					go func() {
						defer par.Release()
						x := us.indexes[defQueryIndexName]
						if err := prepareIndex(us.fs, defQueryIndexName, x); err != nil {
							par.Error(err)
							return
						}
						unitDefQueryIndexesLock.Lock()
						defer unitDefQueryIndexesLock.Unlock()
						unitDefQueryIndexes[u] = x.(*defQueryIndex)
					}()
				}
				getUnitDefQueryIndexesErr = par.Wait()
			}
			if unitDefQueryIndexes == nil {
				unitDefQueryIndexes = map[unit.ID2]*defQueryIndex{}
			}
		})
		return unitDefQueryIndexes, getUnitDefQueryIndexesErr
	}

	par := parallel.NewRun(runtime.GOMAXPROCS(0))
	for name_, x_ := range xs {
		name, x := name_, x_
		par.Acquire()
		go func() {
			defer par.Release()
			switch x := x.(type) {
			case unitIndexBuilder:
				units, err := getUnits()
				if err != nil {
					par.Error(err)
					return
				}
				if err := x.Build(units); err != nil {
					par.Error(err)
					return
				}
			case unitRefIndexBuilder:
				unitRefIndexes, err := getUnitRefIndexes()
				if err != nil {
					par.Error(err)
					return
				}
				if err := x.Build(unitRefIndexes); err != nil {
					par.Error(err)
					return
				}
			case defQueryTreeIndexBuilder:
				unitDefQueryIndexes, err := getUnitDefQueryIndexes()
				if err != nil {
					par.Error(err)
					return
				}
				if err := x.Build(unitDefQueryIndexes); err != nil {
					par.Error(err)
					return
				}
			default:
				par.Error(fmt.Errorf("don't know how to build index %q of type %T", name, x))
				return
			}
			if x, ok := x.(persistedIndex); ok {
				if err := writeIndex(s.fs, name, x); err != nil {
					par.Error(err)
					return
				}
			}
		}()
	}
	return par.Wait()
}

func (s *indexedTreeStore) statIndex(name string) (os.FileInfo, error) {
	return statIndex(s.fs, name)
}

// An indexedUnitStore is a VFS-backed unit store that generates
// indexes to provide efficient lookups.
//
// It wraps a fsUnitStore and intercepts calls to Def, Defs,
// Refs, etc., using its indexes to satisfy those queries efficiently
// where possible. Otherwise it invokes the underlying store to
// perform a full scan.
type indexedUnitStore struct {
	// indexes is all of the indexes that should be built, written,
	// and read from. It contains all indexes for all types of data
	// (e.g., def indexes, ref indexes, etc.).
	indexes map[string]Index

	*fsUnitStore
}

var _ interface {
	UnitStore
	indexedStore
} = (*indexedUnitStore)(nil)

// newIndexedUnitStore creates a new indexed unit store that stores
// data and indexes in fs.
func newIndexedUnitStore(fs rwvfs.FileSystem, label string) UnitStoreImporter {
	return &indexedUnitStore{
		indexes: map[string]Index{
			"path_to_def":      &defPathIndex{},
			"file_to_refs":     &refFileIndex{},
			defToRefsIndexName: &defRefsIndex{},
			defQueryIndexName:  &defQueryIndex{f: defQueryFilter},
		},
		fsUnitStore: &fsUnitStore{fs: fs, label: label},
	}
}

const (
	defToRefsIndexName = "def_to_refs"
	defQueryIndexName  = "def_query"
	indexFilename      = "%s.idx"
)

func (s *indexedUnitStore) Defs(fs ...DefFilter) ([]*graph.Def, error) {
	// If there's a defOffsetsFilter, that'll be faster than
	// consulting an index (since it already gives us the byte
	// offsets).
	if hasDefOffsetsFilter := getDefOffsetsFilter(fs) != nil; !hasDefOffsetsFilter {
		// Try to find an index that covers this query.
		if xname, bx := bestCoverageIndex(s.indexes, fs, isDefIndex); bx != nil {
			if err := prepareIndex(s.fs, xname, bx); err != nil {
				return nil, err
			}
			vlog.Printf("indexedUnitStore.Defs(%v): Found covering index %q (%v).", fs, xname, bx)
			ofs, err := bx.(defIndex).Defs(fs...)
			if err != nil {
				return nil, err
			}
			return s.defsAtOffsets(ofs, fs)
		}
	}

	// Fall back to full scan.
	return s.fsUnitStore.Defs(fs...)
}

// Refs implements UnitStore.
func (s *indexedUnitStore) Refs(fs ...RefFilter) ([]*graph.Ref, error) {
	// Try to find an index that covers this query.
	if xname, bx := bestCoverageIndex(s.indexes, fs, isRefIndex); bx != nil {
		if err := prepareIndex(s.fs, xname, bx); err != nil {
			return nil, err
		}
		vlog.Printf("indexedUnitStore.Refs(%v): Found covering index %q (%v).", fs, xname, bx)
		switch bx := bx.(type) {
		case refIndexByteRanges:
			brs, err := bx.Refs(fs...)
			if err != nil {
				return nil, err
			}
			return s.refsAtByteRanges(brs, fs)
		case refIndexByteOffsets:
			ofs, err := bx.Refs(fs...)
			if err != nil {
				return nil, err
			}
			return s.refsAtOffsets(ofs, fs)
		}
	}

	// Fall back to full scan.
	return s.fsUnitStore.Refs(fs...)
}

// Import calls to the underlying fsUnitStore to write the def
// and ref data files. It also builds and writes the indexes.
func (s *indexedUnitStore) Import(data graph.Output) error {
	cleanForImport(&data, "", "", "")

	var defOfs, refOfs byteOffsets
	var refFBRs fileByteRanges

	var err error

	// Complete each write operation serially which eliminates
	// a lock contention bug on slower underlying file stores
	defOfs, err = s.fsUnitStore.writeDefs(data.Defs)
	if err != nil {
		return err
	}
	refFBRs, refOfs, err = s.fsUnitStore.writeRefs(data.Refs)
	if err != nil {
		return err
	}
	if err := s.buildIndexes(s.Indexes(), &data, defOfs, refFBRs, refOfs); err != nil {
		return err
	}

	return nil
}

func (s *indexedUnitStore) Indexes() map[string]Index { return s.indexes }

func (s *indexedUnitStore) BuildIndex(name string, x Index) error {
	return s.buildIndexes(map[string]Index{name: x}, nil, nil, nil, nil)
}

func (s *indexedUnitStore) readIndex(name string, x persistedIndex) error {
	return readIndex(s.fs, name, x)
}

func (s *indexedUnitStore) buildIndexes(xs map[string]Index, data *graph.Output, defOfs byteOffsets, refFBRs fileByteRanges, refOfs byteOffsets) error {
	var defs []*graph.Def
	var refs []*graph.Ref
	if data != nil {
		// Allow us to distinguish between empty (empty slice) and not-yet-fetched (nil).
		defs = data.Defs
		if defs == nil {
			defs = []*graph.Def{}
		}
		refs = data.Refs
		if refs == nil {
			refs = []*graph.Ref{}
		}
	}

	var getDefsErr error
	var getDefsOnce sync.Once
	getDefs := func() ([]*graph.Def, byteOffsets, error) {
		getDefsOnce.Do(func() {
			// Don't refetch if passed in as arg or if getData was
			// already called.
			if defs == nil {
				defs, defOfs, getDefsErr = s.fsUnitStore.readDefs()
			}
			if defs == nil {
				defs = []*graph.Def{}
			}
		})
		return defs, defOfs, getDefsErr
	}

	var getRefsErr error
	var getRefsOnce sync.Once
	getRefs := func() ([]*graph.Ref, fileByteRanges, byteOffsets, error) {
		getRefsOnce.Do(func() {
			// Don't refetch if passed in as arg or if getData was
			// already called.
			if refs == nil {
				refs, refFBRs, refOfs, getRefsErr = s.fsUnitStore.readRefs()
			}
			if refs == nil {
				refs = []*graph.Ref{}
			}
		})
		return refs, refFBRs, refOfs, getRefsErr
	}

	par := parallel.NewRun(runtime.GOMAXPROCS(0))
	for name_, x_ := range xs {
		name, x := name_, x_
		par.Acquire()
		go func() {
			defer par.Release()
			switch x := x.(type) {
			case defIndexBuilder:
				defs, defOfs, err := getDefs()
				if err != nil {
					par.Error(err)
					return
				}
				if err := x.Build(defs, defOfs); err != nil {
					par.Error(err)
					return
				}
			case refIndexBuilder:
				refs, refFBRs, refOfs, err := getRefs()
				if err != nil {
					par.Error(err)
					return
				}
				if err := x.Build(refs, refFBRs, refOfs); err != nil {
					par.Error(err)
					return
				}
			default:
				par.Error(fmt.Errorf("don't know how to build index %q of type %T", name, x))
				return
			}
			if x, ok := x.(persistedIndex); ok {
				if err := writeIndex(s.fs, name, x); err != nil {
					par.Error(err)
					return
				}
			}
		}()
	}
	return par.Wait()
}

func (s *indexedUnitStore) statIndex(name string) (os.FileInfo, error) {
	return statIndex(s.fs, name)
}

func (s *indexedUnitStore) String() string { return "indexedUnitStore" }

// writeIndex calls x.Write with the index's backing file.
func writeIndex(fs rwvfs.FileSystem, name string, x persistedIndex) (err error) {
	vlog.Printf("%s: writing index...", name)
	f, err := fs.Create(fmt.Sprintf(indexFilename, name))
	if err != nil {
		return err
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()

	w := gzip.NewWriter(f)

	if err := x.Write(w); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	vlog.Printf("%s: done writing index.", name)
	return nil
}

// prepareIndex prepares an index to be used. If it is already Ready,
// nothing happens. If it's not Ready and it's a persistedIndex,
// prepareIndex calls readIndex(fs, name, x). Otherwise an
// *errIndexNotReady is returned.
func prepareIndex(fs rwvfs.FileSystem, name string, x Index) error {
	if x.Ready() {
		return nil
	}
	if x, ok := x.(persistedIndex); ok {
		return readIndex(fs, name, x)
	}
	return &errIndexNotReady{name: name}
}

type errIndexNotReady struct {
	name string
}

func (e *errIndexNotReady) Error() string { return fmt.Sprintf("index not ready: %s", e.name) }

type errIndexNotExist struct {
	name string
	err  error
}

func (e *errIndexNotExist) Error() string {
	return fmt.Sprintf("index %q does not exist: %s", e.name, e.err)
}

// readIndex calls x.Read with the index's backing file.
func readIndex(fs rwvfs.FileSystem, name string, x persistedIndex) (err error) {
	vlog.Printf("%s: reading index...", name)
	var f vfs.ReadSeekCloser
	f, err = fs.Open(fmt.Sprintf(indexFilename, name))
	if err != nil {
		vlog.Printf("%s: failed to read index: %s.", name, err)
		if os.IsNotExist(err) {
			return &errIndexNotExist{name: name, err: err}
		}
		return err
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()

	r, err := gzip.NewReader(f)
	if err != nil {
		return err
	}

	if err := x.Read(r); err != nil {
		return err
	}
	if err := r.Close(); err != nil {
		return err
	}
	vlog.Printf("%s: done reading index.", name)
	return nil
}

// statIndex calls fs.Stat on the index's backing file or dir.
func statIndex(fs rwvfs.FileSystem, name string) (os.FileInfo, error) {
	return fs.Stat(fmt.Sprintf(indexFilename, name))
}

var defQueryFilter = DefFilterFunc(func(def *graph.Def) bool {
	return !def.Local && def.Name != ""
})
