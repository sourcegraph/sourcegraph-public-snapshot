package store

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/neelance/parallel"

	"github.com/kr/fs"
	"golang.org/x/tools/godoc/vfs"

	"sort"

	"sourcegraph.com/sourcegraph/rwvfs"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// useIndexedStore indicates whether the indexed{Unit,Tree}Stores
// should be used. If it's false, only the FS-backed stores are used
// (which requires full scans for all filters).
var (
	noIndex, _      = strconv.ParseBool(os.Getenv("NOINDEX"))
	useIndexedStore = !noIndex
)

// A fsMultiRepoStore is a MultiRepoStore that stores data on a VFS.
type fsMultiRepoStore struct {
	fs rwvfs.WalkableFileSystem
	FSMultiRepoStoreConf
	repoStores
}

var _ MultiRepoStoreImporterIndexer = (*fsMultiRepoStore)(nil)

// NewFSMultiRepoStore creates a new repository store (that can be
// imported into) that is backed by files on a filesystem.
func NewFSMultiRepoStore(fs rwvfs.WalkableFileSystem, conf *FSMultiRepoStoreConf) MultiRepoStoreImporterIndexer {
	if conf == nil {
		conf = &FSMultiRepoStoreConf{}
	}
	if conf.RepoPaths == nil {
		conf.RepoPaths = DefaultRepoPaths
	}

	setCreateParentDirs(fs)
	mrs := &fsMultiRepoStore{fs: fs, FSMultiRepoStoreConf: *conf}
	mrs.repoStores = repoStores{mrs}
	return mrs
}

// FSMultiRepoStoreConf configures an FS-backed multi-repo store. Pass
// it to NewFSMultiRepoStore to construct a new store with the
// specified options.
type FSMultiRepoStoreConf struct {
	// RepoPathConfig specifies where the multi-repo store stores
	// repository data. If nil, DefaultRepoPaths is used, which stores
	// repos at "${REPO}/.srclib-store".
	RepoPaths
}

// getRepo gets a single repo.
func (s *fsMultiRepoStore) getRepo(repo string) (string, error) {
	repoPath := s.fs.Join(s.RepoToPath(repo)...)
	fi, err := s.fs.Stat(repoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	if !fi.Mode().IsDir() {
		return "", nil
	}
	return repo, nil
}

func (s *fsMultiRepoStore) Repos(f ...RepoFilter) ([]string, error) {
	scopeRepos, err := scopeRepos(storeFilters(f))
	if err != nil {
		return nil, err
	}

	var repos []string

	if scopeRepos != nil {
		// Fetch repos individually.
		for _, repo := range scopeRepos {
			repo2, err := s.getRepo(repo)
			if err != nil {
				return nil, err
			}
			if repo2 != "" {
				repos = append(repos, repo2)
			}
		}
	} else {
		// List repos.
		var after string
		var max int
		var allPaths [][]string
		for {
			const maxFetch = 1000
			var numFetch int
			if max == 0 {
				numFetch = maxFetch
			} else {
				numFetch = max - len(allPaths)
				if numFetch > maxFetch {
					numFetch = maxFetch
				}
			}
			if numFetch <= 0 {
				break
			}

			paths, err := s.ListRepoPaths(s.fs, after, numFetch)
			if err != nil {
				return nil, err
			}
			allPaths = append(allPaths, paths...)
			if len(paths) < numFetch {
				break
			}
			after = s.fs.Join(paths[len(paths)-1]...)
		}
		repos = make([]string, len(allPaths))
		for i, path := range allPaths {
			repos[i] = s.PathToRepo(path)
		}
	}

	filteredRepos := make([]string, 0, len(repos))
	for _, repo := range repos {
		if repoFilters(f).SelectRepo(repo) {
			filteredRepos = append(filteredRepos, repo)
		}
	}
	return filteredRepos, nil
}

func (s *fsMultiRepoStore) openRepoStore(repo string) RepoStore {
	subpath := s.fs.Join(s.RepoToPath(repo)...)
	return NewFSRepoStore(rwvfs.Walkable(rwvfs.Sub(s.fs, subpath)))
}

func (s *fsMultiRepoStore) openAllRepoStores() (map[string]RepoStore, error) {
	repos, err := s.Repos()
	if err != nil {
		return nil, err
	}

	rss := make(map[string]RepoStore, len(repos))
	for _, repo := range repos {
		rss[repo] = s.openRepoStore(repo)
	}
	return rss, nil
}

var _ repoStoreOpener = (*fsMultiRepoStore)(nil)

func (s *fsMultiRepoStore) Import(repo, commitID string, unit *unit.SourceUnit, data graph.Output) error {
	if unit != nil {
		cleanForImport(&data, repo, unit.Type, unit.Name)
	}
	subpath := s.fs.Join(s.RepoToPath(repo)...)
	if err := rwvfs.MkdirAll(s.fs, subpath); err != nil {
		return err
	}
	return s.openRepoStore(repo).(RepoImporter).Import(commitID, unit, data)
}

func (s *fsMultiRepoStore) CreateVersion(repo, commitID string) error {
	return s.openRepoStore(repo).(RepoImporter).CreateVersion(commitID)
}

func (s *fsMultiRepoStore) Index(repo, commitID string) error {
	switch rs := s.openRepoStore(repo).(type) {
	case RepoIndexer:
		return rs.Index(commitID)
	}
	return nil
}

func (s *fsMultiRepoStore) String() string { return "fsMultiRepoStore" }

// A fsRepoStore is a RepoStore that stores data on a VFS.
type fsRepoStore struct {
	fs rwvfs.WalkableFileSystem
	treeStores
}

// SrclibStoreDir is the name of the directory under which a RepoStore's data is stored.
const SrclibStoreDir = ".srclib-store"

// NewFSRepoStore creates a new repository store (that can be
// imported into) that is backed by files on a filesystem.
func NewFSRepoStore(fs rwvfs.WalkableFileSystem) RepoStoreImporter {
	setCreateParentDirs(fs)
	rs := &fsRepoStore{fs: fs}
	rs.treeStores = treeStores{rs}
	return rs
}

func (s *fsRepoStore) Versions(f ...VersionFilter) ([]*Version, error) {
	allVersions, err := s.listAllVersions()
	if err != nil {
		return nil, err
	}

	var versions []*Version
	for _, v := range allVersions {
		version := &Version{CommitID: path.Base(v)}
		if versionFilters(f).SelectVersion(version) {
			versions = append(versions, version)
		}
	}
	return versions, nil
}

const (
	versionsDir = "__versions"

	// We can remove version migration in a few months (from 2015 Dec
	// 10).
	enableMigrateVersions = true
)

func (s *fsRepoStore) listAllVersions() ([]string, error) {
	entries, err := s.fs.ReadDir(versionsDir)
	if (os.IsNotExist(err) || (err == nil && len(entries) == 0)) && enableMigrateVersions {
		return s.migrateVersions()
	}
	if err != nil {
		return nil, err
	}
	dirs := make([]string, len(entries))
	for i, e := range entries {
		dirs[i] = e.Name()
	}
	return dirs, nil
}

// migrateVersions is a temporary function that migrates versions from
// being encoded as the dir names under the root, to being encoded as
// the names of empty files in a __versions dir.
//
// migrateVersions returns the list of versions so that listing them
// does not require another operation.
func (s *fsRepoStore) migrateVersions() ([]string, error) {
	versions, err := s.listAllVersions_old()
	if err != nil {
		return nil, err
	}

	if err := s.fs.Mkdir(versionsDir); err != nil && !os.IsExist(err) {
		return nil, err
	}

	writeEmptyVersionFile := func(version string) error {
		f, err := s.fs.Create(s.fs.Join(versionsDir, version))
		if err != nil {
			return err
		}
		f.Write(nil)
		return f.Close()
	}
	for _, v := range versions {
		if err := writeEmptyVersionFile(v); err != nil {
			return nil, fmt.Errorf("during versions migration: %s (migration could not be rolled back, versions list will be incomplete!)", err)
		}
	}

	return versions, nil
}

// listAllVersions_old is the old way of listing versions, where the
// versions were the names of directories. When the storage backend
// was S3 or Google Cloud Storage, this translated into listing key
// prefixes, which is an extremely slow operation. This method is kept
// around to aid in migration.
func (s *fsRepoStore) listAllVersions_old() ([]string, error) {
	entries, err := s.fs.ReadDir(".")
	if err != nil {
		return nil, err
	}
	dirs := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.Name() == versionsDir {
			continue
		}
		dirs = append(dirs, e.Name())
	}
	return dirs, nil
}

func (s *fsRepoStore) Import(commitID string, unit *unit.SourceUnit, data graph.Output) error {
	if unit != nil {
		cleanForImport(&data, "", unit.Type, unit.Name)
	}
	ts := s.newTreeStore(commitID)
	if err := ts.Import(unit, data); err != nil {
		return err
	}
	return nil
}

func (s *fsRepoStore) CreateVersion(commitID string) error {
	if err := s.fs.Mkdir(versionsDir); err != nil && !os.IsExist(err) {
		return err
	}
	f, err := s.fs.Create(s.fs.Join(versionsDir, commitID))
	if err != nil {
		return err
	}
	f.Write(nil)
	return f.Close()
}

func (s *fsRepoStore) Index(commitID string) error {
	if xs, ok := s.newTreeStore(commitID).(*indexedTreeStore); ok {
		return xs.Index()
	}
	return nil // nothing to do
}

func (s *fsRepoStore) treeStoreFS(commitID string) rwvfs.FileSystem {
	return rwvfs.Sub(s.fs, commitID)
}

func (s *fsRepoStore) newTreeStore(commitID string) TreeStoreImporter {
	fs := s.treeStoreFS(commitID)
	if useIndexedStore {
		cacheKey := fs.String()
		return newIndexedTreeStore(fs, cacheKey)
	}
	return newFSTreeStore(fs)
}

func (s *fsRepoStore) openTreeStore(commitID string) TreeStore {
	return s.newTreeStore(commitID)
}

func (s *fsRepoStore) openAllTreeStores() (map[string]TreeStore, error) {
	versions, err := s.listAllVersions()
	if err != nil {
		return nil, err
	}

	tss := make(map[string]TreeStore, len(versions))
	for _, dir := range versions {
		commitID := path.Base(dir)
		tss[commitID] = s.openTreeStore(commitID)
	}
	return tss, nil
}

var _ treeStoreOpener = (*fsRepoStore)(nil)

func (s *fsRepoStore) String() string { return "fsRepoStore" }

// A fsTreeStore is a TreeStore that stores data on a VFS.
type fsTreeStore struct {
	fs rwvfs.FileSystem
	unitStores
}

func newFSTreeStore(fs rwvfs.FileSystem) *fsTreeStore {
	ts := &fsTreeStore{fs: fs}
	ts.unitStores = unitStores{ts}
	return ts
}

var c_fsTreeStore_unitsOpened = &counter{count: new(int64)}

func (s *fsTreeStore) Units(f ...UnitFilter) ([]*unit.SourceUnit, error) {
	var unitFilenames []string

	unitIDs, err := scopeUnits(storeFilters(f))
	if err != nil {
		return nil, err
	}

	if len(unitIDs) > 0 {
		unitFilenames = make([]string, len(unitIDs))
		for i, u := range unitIDs {
			unitFilenames[i] = s.unitFilename(u.Type, u.Name)
		}
	} else {
		unitFilenames, err = s.unitFilenames()
		if err != nil {
			return nil, err
		}
	}

	var units []*unit.SourceUnit
	for _, filename := range unitFilenames {
		c_fsTreeStore_unitsOpened.increment()
		unit, err := s.openUnitFile(filename)
		if err != nil {
			return nil, err
		}
		if unitFilters(f).SelectUnit(unit) {
			units = append(units, unit)
		}
	}
	return units, nil
}

func (s *fsTreeStore) openUnitFile(filename string) (u *unit.SourceUnit, err error) {
	f, err := s.fs.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errUnitNoInit
		}
		return nil, err
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()

	var unit unit.SourceUnit
	_, err = Codec.NewDecoder(f).Decode(&unit)
	return &unit, err
}

func (s *fsTreeStore) unitFilenames() ([]string, error) {
	var files []string
	w := fs.WalkFS(".", rwvfs.Walkable(s.fs))
	for w.Step() {
		if err := w.Err(); err != nil {
			return nil, err
		}
		fi := w.Stat()
		if fi.Mode().IsRegular() && strings.HasSuffix(fi.Name(), unitFileSuffix) {
			files = append(files, filepath.ToSlash(w.Path()))
		}
	}
	return files, nil
}

func (s *fsTreeStore) unitFilename(unitType, unit string) string {
	return path.Join(unit, unitType+unitFileSuffix)
}

const unitFileSuffix = ".unit.json"

func (s *fsTreeStore) Import(u *unit.SourceUnit, data graph.Output) (err error) {
	if u == nil {
		return rwvfs.MkdirAll(s.fs, ".")
	}

	unitFilename := s.unitFilename(u.Type, u.Name)
	if err := rwvfs.MkdirAll(s.fs, path.Dir(unitFilename)); err != nil {
		return err
	}
	f, err := s.fs.Create(unitFilename)
	if err != nil {
		return err
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()
	if _, err := Codec.NewEncoder(f).Encode(u); err != nil {
		return err
	}

	dir := strings.TrimSuffix(unitFilename, unitFileSuffix)
	if err := rwvfs.MkdirAll(s.fs, dir); err != nil {
		return err
	}
	cleanForImport(&data, "", u.Type, u.Name)
	return s.openUnitStore(unit.ID2{Type: u.Type, Name: u.Name}).(UnitStoreImporter).Import(data)
}

func (s *fsTreeStore) openUnitStore(u unit.ID2) UnitStore {
	filename := s.unitFilename(u.Type, u.Name)
	dir := strings.TrimSuffix(filename, unitFileSuffix)
	if useIndexedStore {
		return newIndexedUnitStore(rwvfs.Sub(s.fs, dir), u.String())
	}
	return &fsUnitStore{fs: rwvfs.Sub(s.fs, dir), label: u.String()}
}

func (s *fsTreeStore) openAllUnitStores() (map[unit.ID2]UnitStore, error) {
	unitFiles, err := s.unitFilenames()
	if err != nil {
		return nil, err
	}

	uss := make(map[unit.ID2]UnitStore, len(unitFiles))
	for _, unitFile := range unitFiles {
		// TODO(sqs): duplicated code both here and in openUnitStore
		// for "dir" and "u".
		dir := strings.TrimSuffix(unitFile, unitFileSuffix)
		u := unit.ID2{Type: path.Base(dir), Name: path.Dir(dir)}
		uss[u] = s.openUnitStore(u)
	}
	return uss, nil
}

var _ unitStoreOpener = (*fsTreeStore)(nil)

func (s *fsTreeStore) String() string { return "fsTreeStore" }

// A fsUnitStore is a UnitStore that stores data on a VFS.
//
// It is typically wrapped by an indexedUnitStore, which provides fast
// responses to indexed queries and passes non-indexed queries through
// to this underlying fsUnitStore.
type fsUnitStore struct {
	// fs is the filesystem where data (and indexes, if
	// fsUnitStore is wrapped by an indexedUnitStore) are
	// written to and read from. The store may create multiple files
	// and arbitrary directory trees in fs (for indexes, etc.).
	fs rwvfs.FileSystem

	label string // a human-readable label (included in String() output)
}

const (
	unitDefsFilename = "def.dat"
	unitRefsFilename = "ref.dat"
)

func (s *fsUnitStore) Defs(fs ...DefFilter) (defs []*graph.Def, err error) {
	if f := getDefOffsetsFilter(fs); f != nil {
		return s.defsAtOffsets(byteOffsets(f), fs)
	}

	vlog.Printf("%s: reading defs with filters %v...", s, fs)
	f, err := s.fs.Open(unitDefsFilename)
	if err != nil {
		return nil, err
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()

	dec := Codec.NewDecoder(f)
	for {
		def := &graph.Def{}
		if _, err := dec.Decode(def); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if DefFilters(fs).SelectDef(def) {
			defs = append(defs, def)
		}
	}
	for _, filter := range fs {
		if dSort, ok := filter.(DefsSorter); ok {
			dSort.DefsSort(defs)
			break
		}
	}
	vlog.Printf("%s: read %v defs with filters %v.", s, len(defs), fs)
	return defs, nil
}

// defsAtOffsets reads the defs at the given serialized byte offsets
// from the def data file and returns them in arbitrary order.
func (s *fsUnitStore) defsAtOffsets(ofs byteOffsets, fs []DefFilter) (defs []*graph.Def, err error) {
	vlog.Printf("%s: reading defs at %d offsets with filters %v...", s, len(ofs), fs)
	f, err := openFetcherOrOpen(s.fs, unitDefsFilename)
	if err != nil {
		return nil, err
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()

	ffs := DefFilters(fs)

	p := parFetches(s.fs, fs)
	if p == 0 {
		return nil, nil
	}

	var defsLock sync.Mutex
	par := parallel.NewRun(p)
	for _, ofs_ := range ofs {
		ofs := ofs_
		par.Acquire()
		go func() {
			defer par.Release()

			if _, moreOK := LimitRemaining(fs); !moreOK {
				return
			}

			// Guess how many bytes this def is. The s3vfs (if that's the
			// VFS impl in use) will autofetch beyond that if needed.
			const byteEstimate = 2 * decodeBufSize
			r, err := rangeReader(s.fs, unitDefsFilename, f, ofs, byteEstimate)
			if err != nil {
				par.Error(err)
				return
			}
			dec := Codec.NewDecoder(r)
			var def graph.Def
			if _, err := dec.Decode(&def); err != nil {
				par.Error(err)
				return
			}
			if ffs.SelectDef(&def) {
				defsLock.Lock()
				defs = append(defs, &def)
				defsLock.Unlock()
			}
		}()
	}
	if err := par.Wait(); err != nil {
		return defs, err
	}
	sort.Sort(graph.Defs(defs))
	vlog.Printf("%s: read %v defs at %d offsets with filters %v.", s, len(defs), len(ofs), fs)
	return defs, nil
}

// readDefs reads all defs from the def data file and returns them
// along with their serialized byte offsets.
func (s *fsUnitStore) readDefs() (defs []*graph.Def, ofs byteOffsets, err error) {
	vlog.Printf("%s: reading defs and byte offsets...", s)
	f, err := s.fs.Open(unitDefsFilename)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()

	n := uint64(0)
	dec := Codec.NewDecoder(f)
	for {
		var def graph.Def
		o, err := dec.Decode(&def)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, err
		}

		ofs = append(ofs, int64(n))
		defs = append(defs, &def)

		n += o
	}
	vlog.Printf("%s: read %d defs and byte ranges.", s, len(defs))
	return defs, ofs, nil
}

func (s *fsUnitStore) Refs(fs ...RefFilter) (refs []*graph.Ref, err error) {
	vlog.Printf("%s: reading refs with filters %v...", s, fs)
	f, err := s.fs.Open(unitRefsFilename)
	if err != nil {
		return nil, err
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()

	dec := Codec.NewDecoder(f)
	for {
		var ref graph.Ref
		if _, err := dec.Decode(&ref); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if refFilters(fs).SelectRef(&ref) {
			refs = append(refs, &ref)
		}
	}
	vlog.Printf("%s: read %d refs with filters %v.", s, len(refs), fs)
	return refs, nil
}

// refsAtByteRanges reads the refs at the given serialized byte ranges
// from the ref data file and returns them in arbitrary order.
func (s *fsUnitStore) refsAtByteRanges(brs []byteRanges, fs []RefFilter) (refs []*graph.Ref, err error) {
	vlog.Printf("%s: reading refs at %d byte ranges with filters %v...", s, len(brs), fs)
	f, err := openFetcherOrOpen(s.fs, unitRefsFilename)
	if err != nil {
		return nil, err
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()

	ffs := refFilters(fs)

	p := parFetches(s.fs, fs)
	if p == 0 {
		return nil, nil
	}

	// See how many bytes we need to read to get the refs in all
	// byteRanges.
	readLengths := make([]int64, len(brs))
	totalRefs := 0
	for i, br := range brs {
		var n int64
		for _, b := range br[1:] {
			n += b
			totalRefs++
		}
		readLengths[i] = n
	}

	var refsLock sync.Mutex
	par := parallel.NewRun(p)
	for i_, br_ := range brs {
		i, br := i_, br_
		par.Acquire()
		go func() {
			defer par.Release()

			if _, moreOK := LimitRemaining(fs); !moreOK {
				return
			}

			r, err := rangeReader(s.fs, unitRefsFilename, f, br.start(), readLengths[i])
			if err != nil {
				par.Error(err)
				return
			}
			dec := Codec.NewDecoder(r)
			for range br[1:] {
				var ref graph.Ref
				if _, err := dec.Decode(&ref); err != nil {
					par.Error(err)
					return
				}
				if ffs.SelectRef(&ref) {
					refsLock.Lock()
					refs = append(refs, &ref)
					refsLock.Unlock()
				}
			}
		}()
	}
	if err := par.Wait(); err != nil {
		return refs, err
	}
	sort.Sort(refsByFileStartEnd(refs))
	vlog.Printf("%s: read %d refs at %d byte ranges with filters %v.", s, len(refs), len(brs), fs)
	return refs, nil
}

// refsAtOffsets reads the refs at the given serialized byte offsets
// from the ref data file and returns them in arbitrary order.
func (s *fsUnitStore) refsAtOffsets(ofs byteOffsets, fs []RefFilter) (refs []*graph.Ref, err error) {
	vlog.Printf("%s: reading refs at %d offsets with filters %v...", s, len(ofs), fs)
	f, err := openFetcherOrOpen(s.fs, unitRefsFilename)
	if err != nil {
		return nil, err
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()

	ffs := refFilters(fs)

	p := parFetches(s.fs, fs)
	if p == 0 {
		return nil, nil
	}

	var refsLock sync.Mutex
	par := parallel.NewRun(p)
	for _, ofs_ := range ofs {
		ofs := ofs_
		par.Acquire()
		go func() {
			defer par.Release()

			if _, moreOK := LimitRemaining(fs); !moreOK {
				return
			}

			// Guess how many bytes this ref is. The s3vfs (if that's the
			// VFS impl in use) will autofetch beyond that if needed.
			const byteEstimate = decodeBufSize
			r, err := rangeReader(s.fs, unitRefsFilename, f, ofs, byteEstimate)
			if err != nil {
				par.Error(err)
				return
			}
			dec := Codec.NewDecoder(r)
			var ref graph.Ref
			if _, err := dec.Decode(&ref); err != nil {
				par.Error(err)
				return
			}
			if ffs.SelectRef(&ref) {
				refsLock.Lock()
				refs = append(refs, &ref)
				refsLock.Unlock()
			}
		}()
	}
	if err := par.Wait(); err != nil {
		return refs, err
	}
	sort.Sort(refsByFileStartEnd(refs))
	vlog.Printf("%s: read %v refs at %d offsets with filters %v.", s, len(refs), len(ofs), fs)
	return refs, nil
}

const maxNetPar = 4

// parFetches returns the number of parallel fetches that should be
// attempted given the VFS and filters.
func parFetches(fs rwvfs.FileSystem, filters interface{}) int {
	// It's almost always faster to read local files serially
	// (FetcherOpener is currently only implemented by network VFSs).
	if _, ok := fs.(rwvfs.FetcherOpener); !ok {
		return 1
	}

	n, moreOK := LimitRemaining(filters)
	if moreOK {
		if n == 0 {
			return maxNetPar
		}
		return min(maxNetPar, n)
	}
	return 0
}

// openFetcher calls fs.OpenFetcher if it implemented the
// FetcherOpener interface; otherwise it calls fs.Open.
func openFetcherOrOpen(fs rwvfs.FileSystem, name string) (vfs.ReadSeekCloser, error) {
	if fo, ok := fs.(rwvfs.FetcherOpener); ok {
		return fo.OpenFetcher(name)
	}
	return fs.Open(name)
}

// rangeReader calls ioutil.ReadAll on the given byte range [start, n). It uses
// optimizations for different kinds of VFSs.
func rangeReader(fs rwvfs.FileSystem, name string, f io.ReadSeeker, start, n int64) (io.Reader, error) {
	if fs, ok := fs.(rwvfs.FetcherOpener); ok {
		// Clone f so we can parallelize it.
		var err error
		f, err = fs.OpenFetcher(name)
		if err != nil {
			return nil, err
		}
		if err := f.(rwvfs.Fetcher).Fetch(start, start+n); err != nil {
			return nil, err
		}
	}
	if _, err := f.Seek(start, 0); err != nil {
		return nil, err
	}
	return f, nil
}

// readDefs reads all defs from the def data file and returns them
// along with their serialized byte offsets.
func (s *fsUnitStore) readRefs() (refs []*graph.Ref, fbrs fileByteRanges, ofs byteOffsets, err error) {
	vlog.Println("fsUnitStore: reading all refs and byte ranges...")
	f, err := s.fs.Open(unitRefsFilename)
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()

	o := int64(0)
	dec := Codec.NewDecoder(f)
	fbrs = fileByteRanges{}
	lastFile := ""
	for {
		var ref graph.Ref
		n, err := dec.Decode(&ref)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, nil, err
		}

		ofs = append(ofs, o)

		var lastFileRefStartOffset int64
		if brs, present := fbrs[ref.File]; present {
			lastFileRefStartOffset = brs[len(brs)-1]
		}
		if lastFile != "" && ref.File != lastFile {
			fbrs[lastFile] = append(fbrs[lastFile], o-lastFileRefStartOffset)
		}
		fbrs[ref.File] = append(fbrs[ref.File], o-lastFileRefStartOffset)
		refs = append(refs, &ref)
		lastFile = ref.File

		o += int64(n)
	}
	if lastFile != "" {
		var lastFileRefStartOffset int64
		if brs, present := fbrs[lastFile]; present {
			lastFileRefStartOffset = brs[len(brs)-1]
		}
		fbrs[lastFile] = append(fbrs[lastFile], o-lastFileRefStartOffset)
	}
	vlog.Printf("%s: read %d refs and byte ranges.", s, len(refs))
	return refs, fbrs, ofs, nil

}

func (s *fsUnitStore) Import(data graph.Output) error {
	cleanForImport(&data, "", "", "")
	if _, err := s.writeDefs(data.Defs); err != nil {
		return err
	}
	if _, _, err := s.writeRefs(data.Refs); err != nil {
		return err
	}
	return nil
}

// writeDefs writes the def data file. It also tracks (in ofs) the
// serialized byte offset where each def's serialized representation
// begins (which is used during index construction).
func (s *fsUnitStore) writeDefs(defs []*graph.Def) (ofs byteOffsets, err error) {
	vlog.Printf("%s: writing %d defs...", s, len(defs))
	f, err := s.fs.Create(unitDefsFilename)
	if err != nil {
		return nil, err
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()

	bw := bufio.NewWriter(f)
	enc := Codec.NewEncoder(bw)
	ofs = make(byteOffsets, len(defs))
	var o uint64 // number of bytes read
	for i, def := range defs {
		ofs[i] = int64(o)
		n, err := enc.Encode(def)
		if err != nil {
			return nil, err
		}
		o += n
	}
	if err := bw.Flush(); err != nil {
		return nil, err
	}
	vlog.Printf("%s: done writing %d defs.", s, len(defs))
	return ofs, nil
}

// writeDefs writes the ref data file.
func (s *fsUnitStore) writeRefs(refs []*graph.Ref) (fbr fileByteRanges, ofs byteOffsets, err error) {
	vlog.Printf("%s: writing %d refs...", s, len(refs))
	f, err := s.fs.Create(unitRefsFilename)
	if err != nil {
		return nil, ofs, err
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()

	// Sort refs by file and start byte so that we can use streaming
	// reads to efficiently read in all of the refs that exist in a
	// file.
	t0 := time.Now()
	sort.Sort(refsByFileStartEnd(refs))
	if d := time.Since(t0); d > time.Millisecond*200 {
		vlog.Printf("%s: sorting %d refs took %s.", s, len(refs), d)
	}

	bw := bufio.NewWriter(f)
	enc := Codec.NewEncoder(bw)
	var o uint64
	fbr = fileByteRanges{}
	ofs = make(byteOffsets, len(refs))
	lastFile := ""
	lastFileByteRanges := byteRanges{}
	for i, ref := range refs {
		ofs[i] = int64(o)

		if lastFile != ref.File {
			if lastFile != "" {
				fbr[lastFile] = lastFileByteRanges
			}
			lastFile = ref.File
			lastFileByteRanges = byteRanges{int64(o)}
		}
		before := o
		n, err := enc.Encode(ref)
		if err != nil {
			return nil, ofs, err
		}
		o += n

		// Record the byte length of this encoded ref.
		lastFileByteRanges = append(lastFileByteRanges, int64(o-before))
	}
	if lastFile != "" {
		fbr[lastFile] = lastFileByteRanges
	}
	if err := bw.Flush(); err != nil {
		return nil, ofs, err
	}
	vlog.Printf("%s: done writing %d refs.", s, len(refs))
	return fbr, ofs, nil
}

func (s *fsUnitStore) String() string { return fmt.Sprintf("fsUnitStore(%v)", s.label) }

// countingWriter wraps an io.Writer, counting the number of bytes
// written.

func setCreateParentDirs(fs rwvfs.FileSystem) {
	type createParents interface {
		CreateParentDirs(bool)
	}
	if fs, ok := fs.(createParents); ok {
		fs.CreateParentDirs(true)
	}
}
