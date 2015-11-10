package hg

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/beyang/hgo"
	hg_changelog "github.com/beyang/hgo/changelog"
	hg_revlog "github.com/beyang/hgo/revlog"
	hg_store "github.com/beyang/hgo/store"
	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/hgcmd"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/internal"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/util"
	"sourcegraph.com/sqs/pbtypes"
)

func init() {
	// Overwrite the hg opener to return repositories that use the
	// faster native-Go hg implementation.
	vcs.RegisterOpener("hg", func(dir string) (vcs.Repository, error) {
		return Open(dir)
	})
}

type Repository struct {
	*hgcmd.Repository
	u           *hgo.Repository
	st          *hg_store.Store
	cl          *hg_revlog.Index
	allTags     *hgo.Tags
	branchHeads *hgo.BranchHeads
}

func Open(dir string) (*Repository, error) {
	r, err := hgo.OpenRepository(dir)
	if err != nil {
		return nil, err
	}

	st := r.NewStore()
	cl, err := st.OpenChangeLog()
	if err != nil {
		return nil, err
	}

	globalTags, allTags := r.Tags()
	globalTags.Sort()
	allTags.Sort()
	allTags.Add("tip", cl.Tip().Id().Node())

	bh, err := r.BranchHeads()
	if err != nil {
		return nil, err
	}

	cr, err := hgcmd.Open(dir)
	if err != nil {
		return nil, err
	}

	return &Repository{cr, r, st, cl, allTags, bh}, nil
}

func (r *Repository) ResolveRevision(spec string) (vcs.CommitID, error) {
	if id, err := r.ResolveBranch(spec); err == nil {
		return id, nil
	}
	if id, err := r.ResolveTag(spec); err == nil {
		return id, nil
	}

	rec, err := r.parseRevisionSpec(spec).Lookup(r.cl)
	if err != nil {
		if err == hg_revlog.ErrRevNotFound || err == hex.ErrLength {
			return "", vcs.ErrRevisionNotFound
		}
		return "", err
	}
	return vcs.CommitID(hex.EncodeToString(rec.Id())), nil
}

func (r *Repository) ResolveTag(name string) (vcs.CommitID, error) {
	if id, ok := r.allTags.IdByName[name]; ok {
		return vcs.CommitID(id), nil
	}
	return "", vcs.ErrTagNotFound
}

func (r *Repository) ResolveBranch(name string) (vcs.CommitID, error) {
	if id, ok := r.branchHeads.IdByName[name]; ok {
		return vcs.CommitID(id), nil
	}
	return "", vcs.ErrBranchNotFound
}

func (r *Repository) Branches(opt vcs.BranchesOptions) ([]*vcs.Branch, error) {
	if opt.ContainsCommit != "" {
		return nil, fmt.Errorf("vcs.BranchesOptions.ContainsCommit option not implemented")
	}

	bs := make([]*vcs.Branch, len(r.branchHeads.IdByName))
	i := 0
	for name, id := range r.branchHeads.IdByName {
		bs[i] = &vcs.Branch{Name: name, Head: vcs.CommitID(id)}
		i++
	}
	sort.Sort(vcs.Branches(bs))
	return bs, nil
}

func (r *Repository) Tags() ([]*vcs.Tag, error) {
	ts := make([]*vcs.Tag, len(r.allTags.IdByName))
	i := 0
	for name, id := range r.allTags.IdByName {
		ts[i] = &vcs.Tag{Name: name, CommitID: vcs.CommitID(id)}
		i++
	}
	sort.Sort(vcs.Tags(ts))
	return ts, nil
}

func (r *Repository) getRec(id vcs.CommitID) (*hg_revlog.Rec, error) {
	rec, err := hg_revlog.NodeIdRevSpec(id).Lookup(r.cl)
	if err == hg_revlog.ErrRevNotFound {
		err = vcs.ErrCommitNotFound
	}
	return rec, err
}

func (r *Repository) GetCommit(id vcs.CommitID) (*vcs.Commit, error) {
	rec, err := r.getRec(id)
	if err != nil {
		return nil, err
	}
	return r.makeCommit(rec)
}

func (r *Repository) Commits(opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error) {
	rec, err := r.getRec(opt.Head)
	if err != nil {
		return nil, 0, err
	}

	var commits []*vcs.Commit
	total := uint(0)
	for ; ; rec = rec.Prev() {
		if total >= opt.Skip && (opt.N == 0 || uint(len(commits)) < opt.N) {
			c, err := r.makeCommit(rec)
			if err != nil {
				return nil, 0, err
			}
			commits = append(commits, c)
		}
		total++

		if rec.IsStartOfBranch() {
			break
		}
		// If we want total, keep going until the end.
		if !opt.NoTotal {
			continue
		}
		// Otherwise return once N has been satisfied.
		if opt.N != 0 && uint(len(commits)) >= opt.N {
			break
		}
	}
	if opt.NoTotal {
		total = 0
	}
	return commits, total, nil
}

func (r *Repository) makeCommit(rec *hg_revlog.Rec) (*vcs.Commit, error) {
	fb := hg_revlog.NewFileBuilder()
	ce, err := hg_changelog.BuildEntry(rec, fb)
	if err != nil {
		return nil, err
	}

	addr, err := mail.ParseAddress(ce.Committer)
	if err != nil {
		// This occurs when the commit author specifier is
		// malformed. Fall back to just using the whole committer
		// string as the name.
		addr = &mail.Address{
			Name:    ce.Committer,
			Address: "",
		}
	}

	var parents []vcs.CommitID
	if !rec.IsStartOfBranch() {
		if p := rec.Parent(); p != nil {
			parents = append(parents, vcs.CommitID(hex.EncodeToString(rec.Parent().Id())))
		}
		if rec.Parent2Present() {
			parents = append(parents, vcs.CommitID(hex.EncodeToString(rec.Parent2().Id())))
		}
	}

	return &vcs.Commit{
		ID:      vcs.CommitID(ce.Id),
		Author:  vcs.Signature{addr.Name, addr.Address, pbtypes.NewTimestamp(ce.Date)},
		Message: ce.Comment,
		Parents: parents,
	}, nil
}

func (r *Repository) FileSystem(at vcs.CommitID) (vfs.FileSystem, error) {
	rec, err := r.getRec(at)
	if err != nil {
		return nil, err
	}

	return &hgFSNative{
		dir:  r.Dir,
		at:   hg_revlog.FileRevSpec(rec.FileRev()),
		repo: r.u,
		st:   r.st,
		cl:   r.cl,
		fb:   hg_revlog.NewFileBuilder(),
	}, nil
}

func (r *Repository) parseRevisionSpec(s string) hg_revlog.RevisionSpec {
	if s == "" {
		s = "tip"
		// TODO(sqs): determine per-repository default branch name (not always "default"?)
	}
	if s == "tip" {
		return hg_revlog.TipRevSpec{}
	}
	if s == "null" {
		return hg_revlog.NullRevSpec{}
	}
	if id, ok := r.allTags.IdByName[s]; ok {
		s = id
	} else if i, err := strconv.Atoi(s); err == nil {
		return hg_revlog.FileRevSpec(i)
	}

	return hg_revlog.NodeIdRevSpec(s)
}

type hgFSNative struct {
	dir  string
	at   hg_revlog.FileRevSpec
	repo *hgo.Repository
	st   *hg_store.Store
	cl   *hg_revlog.Index
	fb   *hg_revlog.FileBuilder
}

func (fs *hgFSNative) manifestEntry(chgId hg_revlog.FileRevSpec, fileName string) (me *hg_store.ManifestEnt, err error) {
	m, err := fs.getManifest(chgId)
	if err != nil {
		return
	}
	me = m.Map()[fileName]
	if me == nil {
		err = ErrFileNotInManifest
	}
	return
}

func (fs *hgFSNative) getManifest(chgId hg_revlog.FileRevSpec) (m hg_store.Manifest, err error) {
	rec, err := chgId.Lookup(fs.cl)
	if err != nil {
		return
	}
	c, err := hg_changelog.BuildEntry(rec, fs.fb)
	if err != nil {
		return
	}

	// st := fs.repo.NewStore()
	mlog, err := fs.st.OpenManifests()
	if err != nil {
		return nil, err
	}

	rec2, err := mlog.LookupRevision(int(c.Linkrev), c.ManifestNode)
	if err != nil {
		return nil, err
	}

	return hg_store.BuildManifest(rec2, fs.fb)
}

func (fs *hgFSNative) getEntry(path string) (*hg_revlog.Rec, *hg_store.ManifestEnt, error) {
	path = filepath.ToSlash(path)
	fileLog, err := fs.st.OpenRevlog(path)
	if err != nil {
		return nil, nil, err
	}

	// Get file entry from manifest (so we can look up the correct record in the file revlog)
	ent, err := fs.manifestEntry(fs.at, path)
	if err != nil {
		return nil, nil, err
	}

	// Lookup record in revlog
	entId, err := ent.Id()
	if err != nil {
		return nil, nil, err
	}
	linkRevSpec := hg_revlog.LinkRevSpec{
		Rev: int(fs.at),
		FindPresent: func(maybeAncestors []*hg_revlog.Rec) (index int, err error) {
			// Find the present file record by matching against the manifest entry.
			//
			// Note: this is necessary due to some edge case where the returned file record nodeid does not match the
			// entry in the manifest. So far, have not been able to repro this outside the Python standard library (try
			// Stat() on the README)
			for a, anc := range maybeAncestors {
				if anc.Id().Eq(entId) {
					return a, nil
				}
			}

			ancIds := make([]hg_revlog.NodeId, len(maybeAncestors))
			for a, anc := range maybeAncestors {
				ancIds[a] = anc.Id()
			}
			return 0, fmt.Errorf("failed to find file record with nodeid matching manifest entry nodeid %v, candidates were %v",
				entId, ancIds)
		},
	}
	rec, err := linkRevSpec.Lookup(fileLog)
	if err != nil {
		// HACK HACK. The above workaround for the Python standard
		// library breaks go-vcs for the file "README" in
		// https://golang.org/x/tools rev
		// 536b79981a09daabecc6bc20b8dd4438a55dc12b. So, if we get an
		// error using the linkRevSpec with the FindPresent func, try
		// again without FindPresent.
		linkRevSpec.FindPresent = nil
		rec, err = linkRevSpec.Lookup(fileLog)
		if err != nil {
			return nil, nil, err
		}
	}
	if rec.FileRev() == -1 {
		return nil, nil, hg_revlog.ErrRevisionNotFound
	}

	if int(rec.Linkrev) == int(fs.at) {
		// The requested revision matches this record, which can be
		// used as a sign that the file exists. (TODO(sqs): original comments
		// say maybe this means the file is NOT existent yet? the word "not" is
		// not there but that seems to be a mistake.)
		return rec, ent, nil
	}

	if !rec.IsLeaf() {
		// There are other records that have the current record as a parent.
		// This means, the file was existent, no need to check the manifest.
		return rec, ent, nil
	}

	return rec, ent, nil
}

func (fs *hgFSNative) Open(name string) (vfs.ReadSeekCloser, error) {
	name = internal.Rel(name)
	rec, _, err := fs.getEntry(name)
	if err != nil {
		return nil, standardizeHgError(err)
	}

	data, err := fs.readFile(rec)
	if err != nil {
		return nil, err
	}
	return util.NopCloser{bytes.NewReader(data)}, nil
}

func (fs *hgFSNative) readFile(rec *hg_revlog.Rec) ([]byte, error) {
	fb := hg_revlog.NewFileBuilder()
	return fb.Build(rec)
}

func (fs *hgFSNative) getModTime() (time.Time, error) {
	r, err := fs.at.Lookup(fs.cl)
	if err != nil {
		return time.Time{}, err
	}

	c, err := hg_changelog.BuildEntry(r, fs.fb)
	if err != nil {
		return time.Time{}, err
	}

	return c.Date, nil
}

func (fs *hgFSNative) Lstat(path string) (os.FileInfo, error) {
	fi, _, err := fs.lstat(path)
	return fi, err
}

func (fs *hgFSNative) lstat(path string) (*util.FileInfo, []byte, error) {
	path = filepath.Clean(internal.Rel(path))

	rec, ent, err := fs.getEntry(path)
	if os.IsNotExist(err) {
		// check if path is a dir (dirs are not in hg's manifest, so we need to
		// hack around to get them).
		fi, err := fs.dirStat(path)
		return fi, nil, err
	}
	if err != nil {
		return nil, nil, standardizeHgError(err)
	}

	fi := fs.fileInfo(ent)

	// read data to determine file size
	data, err := fs.readFile(rec)
	if err != nil {
		return nil, nil, err
	}
	fi.Size_ = int64(len(data))

	return fi, data, nil
}

func (fs *hgFSNative) Stat(path string) (os.FileInfo, error) {
	path = internal.Rel(path)
	fi, data, err := fs.lstat(path)
	if err != nil {
		return nil, err
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		// derefence symlink
		derefPath := string(data)
		fi, err := fs.Lstat(derefPath)
		if err != nil {
			return nil, err
		}

		fi.(*util.FileInfo).Name_ = filepath.Base(path)
		return fi, nil
	}

	return fi, nil
}

// dirStat determines whether a directory exists at path by listing files
// underneath it. If it has files, then it's a directory. We must do it this way
// because hg doesn't track directories in the manifest.
func (fs *hgFSNative) dirStat(path string) (*util.FileInfo, error) {
	mtime, err := fs.getModTime()
	if err != nil {
		return nil, err
	}

	if path == "." {
		return &util.FileInfo{
			Name_:    ".",
			Mode_:    os.ModeDir,
			ModTime_: mtime,
		}, nil
	}

	m, err := fs.getManifest(fs.at)
	if err != nil {
		return nil, err
	}

	dirPrefix := filepath.Clean(path) + "/"
	for _, e := range m {
		if strings.HasPrefix(e.FileName, dirPrefix) {
			return &util.FileInfo{
				Name_:    filepath.Base(path),
				Mode_:    os.ModeDir,
				ModTime_: mtime,
			}, nil
		}
	}

	return nil, os.ErrNotExist
}

func (fs *hgFSNative) fileInfo(ent *hg_store.ManifestEnt) *util.FileInfo {
	var mode os.FileMode

	mtime, err := fs.getModTime()
	if err != nil {
		return nil
	}

	if ent.IsExecutable() {
		mode |= 0111 // +x
	}
	if ent.IsLink() {
		mode |= os.ModeSymlink
	}

	return &util.FileInfo{
		Name_:    filepath.Base(ent.FileName),
		Mode_:    mode,
		ModTime_: mtime,
	}
}

func (fs *hgFSNative) ReadDir(path string) ([]os.FileInfo, error) {
	path = internal.Rel(path)
	m, err := fs.getManifest(fs.at)
	if err != nil {
		return nil, err
	}

	var fis []os.FileInfo
	subdirs := make(map[string]struct{})

	var dirPrefix string
	if path := filepath.Clean(path); path == "." {
		dirPrefix = ""
	} else {
		dirPrefix = path + "/"
	}
	for _, e := range m {
		if !strings.HasPrefix(e.FileName, dirPrefix) {
			continue
		}
		name := strings.TrimPrefix(e.FileName, dirPrefix)
		dir := filepath.Dir(name)
		if dir == "." {
			fis = append(fis, fs.fileInfo(&e))
		} else {
			subdir := strings.SplitN(dir, "/", 2)[0]
			if _, seen := subdirs[subdir]; !seen {
				fis = append(fis, &util.FileInfo{Name_: subdir, Mode_: os.ModeDir})
				subdirs[subdir] = struct{}{}
			}
		}
	}
	return fis, nil
}

func (fs *hgFSNative) String() string {
	return fmt.Sprintf("hg repository %s commit %s (native)", fs.dir, fs.at)
}

func standardizeHgError(err error) error {
	switch err {
	case hg_revlog.ErrRevisionNotFound:
		fallthrough
	case ErrFileNotInManifest:
		return os.ErrNotExist
	default:
		return err
	}
}

var ErrFileNotInManifest = errors.New("file does not exist in given revision")
