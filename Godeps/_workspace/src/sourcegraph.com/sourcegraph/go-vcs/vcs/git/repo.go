package git

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	git2go "github.com/libgit2/git2go"
	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/gitcmd"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/internal"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/util"
	"sourcegraph.com/sqs/pbtypes"
)

func init() {
	// Overwrite the git opener to return repositories that use the
	// faster libgit2 implementation.
	vcs.RegisterOpener("git", func(dir string) (vcs.Repository, error) {
		return Open(dir)
	})
}

type Repository struct {
	*gitcmd.Repository
	u *git2go.Repository

	editLock sync.RWMutex // protects ops that change repository data
}

func (r *Repository) String() string {
	return fmt.Sprintf("git (libgit2) repo at %s", r.Dir)
}

func Open(dir string) (*Repository, error) {
	cr, err := gitcmd.Open(dir)
	if err != nil {
		return nil, err
	}

	u, err := git2go.OpenRepository(dir)
	if err != nil {
		return nil, err
	}
	return &Repository{Repository: cr, u: u}, nil
}

func (r *Repository) ResolveRevision(spec string) (vcs.CommitID, error) {
	r.editLock.RLock()
	defer r.editLock.RUnlock()

	o, err := r.u.RevparseSingle(spec)
	if err != nil {
		if err.Error() == fmt.Sprintf("Revspec '%s' not found.", spec) {
			return "", vcs.ErrRevisionNotFound
		}
		return "", err
	}
	defer o.Free()
	return vcs.CommitID(o.Id().String()), nil
}

func (r *Repository) ResolveRef(name string) (vcs.CommitID, error) {
	r.editLock.RLock()
	defer r.editLock.RUnlock()

	ref, err := r.u.References.Lookup(name)
	if err != nil {
		if e, ok := err.(*git2go.GitError); ok && e.Code == git2go.ErrNotFound {
			return "", vcs.ErrRefNotFound
		}
		return "", err
	}
	commit, err := r.u.LookupCommit(ref.Target())
	if err != nil {
		return "", err
	}
	defer commit.Free()
	return vcs.CommitID(commit.Id().String()), nil
}

func (r *Repository) ResolveBranch(name string) (vcs.CommitID, error) {
	r.editLock.RLock()
	defer r.editLock.RUnlock()

	b, err := r.u.LookupBranch(name, git2go.BranchLocal)
	if err != nil {
		if err.Error() == fmt.Sprintf("Cannot locate local branch '%s'", name) {
			return "", vcs.ErrBranchNotFound
		}
		return "", err
	}
	return vcs.CommitID(b.Target().String()), nil
}

func (r *Repository) ResolveTag(name string) (vcs.CommitID, error) {
	r.editLock.RLock()
	defer r.editLock.RUnlock()

	// TODO(sqs): slow way to iterate through tags because git_tag_lookup is not
	// in git2go yet
	refs, err := r.u.NewReferenceIterator()
	if err != nil {
		return "", err
	}

	for {
		ref, err := refs.Next()
		if err != nil {
			break
		}
		if ref.IsTag() && ref.Shorthand() == name {
			return vcs.CommitID(ref.Target().String()), nil
		}
	}

	return "", vcs.ErrTagNotFound
}

func (r *Repository) Branches(opt vcs.BranchesOptions) ([]*vcs.Branch, error) {
	if opt.ContainsCommit != "" {
		return nil, fmt.Errorf("vcs.BranchesOptions.ContainsCommit option not implemented")
	}

	r.editLock.RLock()
	defer r.editLock.RUnlock()

	refs, err := r.u.NewReferenceIterator()
	if err != nil {
		return nil, err
	}

	var bs []*vcs.Branch
	for {
		ref, err := refs.Next()
		if isErrIterOver(err) {
			break
		}
		if err != nil {
			return nil, err
		}
		if ref.IsBranch() {
			bs = append(bs, &vcs.Branch{Name: ref.Shorthand(), Head: vcs.CommitID(ref.Target().String())})
		}
	}

	sort.Sort(vcs.Branches(bs))
	return bs, nil
}

func (r *Repository) Tags() ([]*vcs.Tag, error) {
	r.editLock.RLock()
	defer r.editLock.RUnlock()

	refs, err := r.u.NewReferenceIterator()
	if err != nil {
		return nil, err
	}

	var ts []*vcs.Tag
	for {
		ref, err := refs.Next()
		if isErrIterOver(err) {
			break
		}
		if err != nil {
			return nil, err
		}
		if ref.IsTag() {
			ts = append(ts, &vcs.Tag{Name: ref.Shorthand(), CommitID: vcs.CommitID(ref.Target().String())})
		}
	}

	sort.Sort(vcs.Tags(ts))
	return ts, nil
}

// getCommit finds and returns the raw git2go Commit. The caller is
// responsible for freeing it (c.Free()).
func (r *Repository) getCommit(id vcs.CommitID) (*git2go.Commit, error) {
	oid, err := git2go.NewOid(string(id))
	if err != nil {
		return nil, err
	}

	c, err := r.u.LookupCommit(oid)
	if err != nil {
		if git2go.IsErrorCode(err, git2go.ErrNotFound) {
			return nil, vcs.ErrCommitNotFound
		}
		return nil, err
	}
	return c, nil
}

func (r *Repository) GetCommit(id vcs.CommitID) (*vcs.Commit, error) {
	r.editLock.RLock()
	defer r.editLock.RUnlock()

	c, err := r.getCommit(id)
	if err != nil {
		return nil, err
	}
	defer c.Free()
	return r.makeCommit(c), nil
}

func (r *Repository) Commits(opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error) {
	r.editLock.RLock()
	defer r.editLock.RUnlock()

	walk, err := r.u.Walk()
	if err != nil {
		return nil, 0, err
	}
	defer walk.Free()

	walk.Sorting(git2go.SortTime)

	oid, err := git2go.NewOid(string(opt.Head))
	if err != nil {
		return nil, 0, err
	}
	if err := walk.Push(oid); err != nil {
		if git2go.IsErrorCode(err, git2go.ErrNotFound) {
			return nil, 0, vcs.ErrCommitNotFound
		}
		return nil, 0, err
	}

	if opt.Base != "" {
		baseOID, err := git2go.NewOid(string(opt.Base))
		if err != nil {
			return nil, 0, err
		}
		if err := walk.Hide(baseOID); err != nil {
			if git2go.IsErrorCode(err, git2go.ErrNotFound) {
				return nil, 0, vcs.ErrCommitNotFound
			}
			return nil, 0, err
		}
	}

	var commits []*vcs.Commit
	total := uint(0)
	err = walk.Iterate(func(c *git2go.Commit) bool {
		if total >= opt.Skip && (opt.N == 0 || uint(len(commits)) < opt.N) {
			commits = append(commits, r.makeCommit(c))
		}
		total++
		// If we want total, keep going until the end.
		if !opt.NoTotal {
			return true
		}
		// Otherwise return once N has been satisfied.
		return (opt.N == 0 || uint(len(commits)) < opt.N)
	})
	if err != nil {
		return nil, 0, err
	}
	if opt.NoTotal {
		total = 0
	}

	return commits, total, nil
}

func (r *Repository) makeCommit(c *git2go.Commit) *vcs.Commit {
	var parents []vcs.CommitID
	if pc := c.ParentCount(); pc > 0 {
		parents = make([]vcs.CommitID, pc)
		for i := 0; i < int(pc); i++ {
			parents[i] = vcs.CommitID(c.ParentId(uint(i)).String())
		}
	}

	au, cm := c.Author(), c.Committer()
	return &vcs.Commit{
		ID:        vcs.CommitID(c.Id().String()),
		Author:    vcs.Signature{au.Name, au.Email, pbtypes.NewTimestamp(au.When)},
		Committer: &vcs.Signature{cm.Name, cm.Email, pbtypes.NewTimestamp(cm.When)},
		Message:   strings.TrimSuffix(c.Message(), "\n"),
		Parents:   parents,
	}
}

var defaultDiffOptions git2go.DiffOptions

func init() {
	var err error
	defaultDiffOptions, err = git2go.DefaultDiffOptions()
	if err != nil {
		log.Fatalf("Failed to load default git (libgit2/git2go) diff options: %s.", err)
	}
	defaultDiffOptions.IdAbbrev = 40
}

func (r *Repository) CrossRepoDiff(base vcs.CommitID, headRepo vcs.Repository, head vcs.CommitID, opt *vcs.DiffOptions) (diff *vcs.Diff, err error) {
	// libgit2 Repository inherits GitRootDir and CrossRepo from its
	// embedded gitcmd.Repository.

	var headDir string // path to head repo on local filesystem
	if headRepo, ok := headRepo.(gitcmd.CrossRepo); ok {
		headDir = headRepo.GitRootDir()
	} else {
		return nil, fmt.Errorf("git cross-repo diff not supported against head repo type %T", headRepo)
	}

	if headDir == r.Dir {
		return r.Diff(base, head, opt)
	}

	r.editLock.Lock()
	defer r.editLock.Unlock()

	rem, err := r.createAndFetchFromAnonRemote(headDir)
	if err != nil {
		return nil, err
	}
	defer rem.Free()

	return r.diffHoldingEditLock(base, head, opt)
}

// createAndFetchFromAnonRemote creates an anonymous git remote and
// fetches from it. The returned remote (if non-nil) should be freed
// (using its Free method) after use.
//
// Callers must hold the r.editLock write lock.
func (r *Repository) createAndFetchFromAnonRemote(repoDir string) (*git2go.Remote, error) {
	name := base64.URLEncoding.EncodeToString([]byte(repoDir))
	rem, err := r.u.Remotes.CreateAnonymous(repoDir)
	if err != nil {
		return nil, err
	}
	if err := rem.Fetch([]string{"+refs/heads/*:refs/remotes/" + name + "/*"}, nil, ""); err != nil {
		rem.Free()
		return nil, err
	}
	return rem, nil
}

func (r *Repository) Diff(base, head vcs.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error) {
	r.editLock.RLock()
	defer r.editLock.RUnlock()
	return r.diffHoldingEditLock(base, head, opt)
}

// diffHoldingLock performs a diff. It must be called while holding
// r.editLock (either as a reader or writer).
func (r *Repository) diffHoldingEditLock(base, head vcs.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error) {
	if opt == nil {
		opt = &vcs.DiffOptions{}
	}

	if opt.ExcludeReachableFromBoth {
		// Not implemented in libgit2 yet, so call gitcmd.
		return r.Repository.Diff(base, head, opt)
	}

	gopt := defaultDiffOptions
	gopt.OldPrefix = opt.OrigPrefix
	gopt.NewPrefix = opt.NewPrefix

	baseCommit, err := r.getCommit(base)
	if err != nil {
		return nil, err
	}
	defer baseCommit.Free()
	baseTree, err := r.u.LookupTree(baseCommit.TreeId())
	if err != nil {
		return nil, err
	}
	defer baseTree.Free()

	headCommit, err := r.getCommit(head)
	if err != nil {
		return nil, err
	}
	defer headCommit.Free()
	headTree, err := r.u.LookupTree(headCommit.TreeId())
	if err != nil {
		return nil, err
	}
	defer headTree.Free()

	if opt != nil {
		if opt.Paths != nil {
			gopt.Pathspec = opt.Paths
		}
	}

	gdiff, err := r.u.DiffTreeToTree(baseTree, headTree, &gopt)
	if err != nil {
		return nil, err
	}
	defer gdiff.Free()

	if opt.DetectRenames {
		findOpts, err := git2go.DefaultDiffFindOptions()
		if err != nil {
			return nil, err
		}
		if err := gdiff.FindSimilar(&findOpts); err != nil {
			return nil, err
		}
	}

	diff := &vcs.Diff{}

	ndeltas, err := gdiff.NumDeltas()
	if err != nil {
		return nil, err
	}
	for i := 0; i < ndeltas; i++ {
		patch, err := gdiff.Patch(i)
		if err != nil {
			return nil, err
		}
		defer patch.Free()

		patchStr, err := patch.String()
		if err != nil {
			return nil, err
		}

		diff.Raw += patchStr
	}
	return diff, nil
}

func (r *Repository) BlameFile(path string, opt *vcs.BlameOptions) ([]*vcs.Hunk, error) {
	r.editLock.RLock()
	defer r.editLock.RUnlock()

	gopt := git2go.BlameOptions{}
	if opt != nil {
		var err error
		if opt.NewestCommit != "" {
			gopt.NewestCommit, err = git2go.NewOid(string(opt.NewestCommit))
			if err != nil {
				return nil, err
			}
		}
		if opt.OldestCommit != "" {
			gopt.OldestCommit, err = git2go.NewOid(string(opt.OldestCommit))
			if err != nil {
				return nil, err
			}
		}
		gopt.MinLine = uint32(opt.StartLine)
		gopt.MaxLine = uint32(opt.EndLine)
	}

	blame, err := r.u.BlameFile(path, &gopt)
	if err != nil {
		return nil, err
	}
	defer blame.Free()

	// Read file contents so we can set hunk byte start and end.
	fs, err := r.FileSystem(vcs.CommitID(gopt.NewestCommit.String()))
	if err != nil {
		return nil, err
	}
	b, err := fs.(*gitFSLibGit2).readFileBytes(path)
	if err != nil {
		return nil, err
	}
	lines := bytes.SplitAfter(b, []byte{'\n'})

	byteOffset := 0
	hunks := make([]*vcs.Hunk, blame.HunkCount())
	for i := 0; i < len(hunks); i++ {
		hunk, err := blame.HunkByIndex(i)
		if err != nil {
			return nil, err
		}

		hunkBytes := 0
		for j := uint16(0); j < hunk.LinesInHunk; j++ {
			hunkBytes += len(lines[j])
		}
		endByteOffset := byteOffset + hunkBytes

		hunks[i] = &vcs.Hunk{
			StartLine: int(hunk.FinalStartLineNumber),
			EndLine:   int(hunk.FinalStartLineNumber + hunk.LinesInHunk),
			StartByte: byteOffset,
			EndByte:   endByteOffset,
			CommitID:  vcs.CommitID(hunk.FinalCommitId.String()),
			Author: vcs.Signature{
				Name:  hunk.FinalSignature.Name,
				Email: hunk.FinalSignature.Email,
				Date:  pbtypes.NewTimestamp(hunk.FinalSignature.When.In(time.UTC)),
			},
		}
		byteOffset = endByteOffset
		lines = lines[hunk.LinesInHunk:]
	}

	return hunks, nil
}

func (r *Repository) MergeBase(a, b vcs.CommitID) (vcs.CommitID, error) {
	r.editLock.RLock()
	defer r.editLock.RUnlock()
	return r.mergeBaseHoldingEditLock(a, b)
}

// mergeBaseHoldingEditLock performs a merge-base. Callers must hold
// the r.editLock (either as a reader or writer).
func (r *Repository) mergeBaseHoldingEditLock(a, b vcs.CommitID) (vcs.CommitID, error) {
	ao, err := git2go.NewOid(string(a))
	if err != nil {
		return "", err
	}
	bo, err := git2go.NewOid(string(b))
	if err != nil {
		return "", err
	}
	mb, err := r.u.MergeBase(ao, bo)
	if err != nil {
		return "", err
	}
	return vcs.CommitID(mb.String()), nil
}

func (r *Repository) CrossRepoMergeBase(a vcs.CommitID, repoB vcs.Repository, b vcs.CommitID) (vcs.CommitID, error) {
	// libgit2 Repository inherits GitRootDir and CrossRepo from its
	// embedded gitcmd.Repository.

	var repoBDir string // path to head repo on local filesystem
	if repoB, ok := repoB.(gitcmd.CrossRepo); ok {
		repoBDir = repoB.GitRootDir()
	} else {
		return "", fmt.Errorf("git cross-repo merge-base not supported against repo type %T", repoB)
	}

	if repoBDir == r.Dir {
		return r.MergeBase(a, b)
	}

	r.editLock.Lock()
	defer r.editLock.Unlock()

	rem, err := r.createAndFetchFromAnonRemote(repoBDir)
	if err != nil {
		return "", err
	}
	defer rem.Free()

	return r.mergeBaseHoldingEditLock(a, b)
}

// TODO(sqs): implement Search using libgit2 (currently falls back to
// gitcmd impl in embedded struct).

func (r *Repository) FileSystem(at vcs.CommitID) (vfs.FileSystem, error) {
	r.editLock.RLock()
	defer r.editLock.RUnlock()

	c, err := r.getCommit(at)
	if err != nil {
		return nil, err
	}
	defer c.Free()

	tree, err := c.Tree()
	if err != nil {
		return nil, err
	}
	return &gitFSLibGit2{r.Dir, c.Id(), at, tree, r.u, &r.editLock}, nil
}

type gitFSLibGit2 struct {
	dir  string
	oid  *git2go.Oid
	at   vcs.CommitID
	tree *git2go.Tree

	repo         *git2go.Repository
	repoEditLock *sync.RWMutex
}

func (fs *gitFSLibGit2) getEntry(path string) (*git2go.TreeEntry, error) {
	path = filepath.Clean(path)
	e, err := fs.tree.EntryByPath(path)
	if err != nil {
		return nil, standardizeLibGit2Error(err)
	}

	return e, nil
}

func (fs *gitFSLibGit2) readFileBytes(name string) ([]byte, error) {
	e, err := fs.getEntry(name)
	if err != nil {
		return nil, err
	}

	switch e.Type {
	case git2go.ObjectBlob:
		b, err := fs.repo.LookupBlob(e.Id)
		if err != nil {
			return nil, err
		}
		defer b.Free()
		return b.Contents(), nil
	case git2go.ObjectCommit:
		// Return empty for a submodule for now.
		return nil, nil
	}
	return nil, fmt.Errorf("read unexpected entry type %q (expected blob or submodule(commit))", e.Type)
}

func (fs *gitFSLibGit2) Open(name string) (vfs.ReadSeekCloser, error) {
	name = internal.Rel(name)

	fs.repoEditLock.RLock()
	defer fs.repoEditLock.RUnlock()

	b, err := fs.readFileBytes(name)
	if err != nil {
		return nil, err
	}
	return util.NopCloser{ReadSeeker: bytes.NewReader(b)}, nil
}

func (fs *gitFSLibGit2) Lstat(path string) (os.FileInfo, error) {
	fs.repoEditLock.RLock()
	defer fs.repoEditLock.RUnlock()

	path = filepath.Clean(internal.Rel(path))

	mtime, err := fs.getModTime()
	if err != nil {
		return nil, err
	}

	if path == "." {
		return &util.FileInfo{Mode_: os.ModeDir, ModTime_: mtime}, nil
	}

	e, err := fs.getEntry(path)
	if err != nil {
		return nil, err
	}

	fi, err := fs.makeFileInfo(path, e)
	if err != nil {
		return nil, err
	}
	fi.ModTime_ = mtime

	return fi, nil
}

func (fs *gitFSLibGit2) Stat(path string) (os.FileInfo, error) {
	fs.repoEditLock.RLock()
	defer fs.repoEditLock.RUnlock()

	path = filepath.Clean(internal.Rel(path))

	mtime, err := fs.getModTime()
	if err != nil {
		return nil, err
	}

	if path == "." {
		return &util.FileInfo{Mode_: os.ModeDir, ModTime_: mtime}, nil
	}

	e, err := fs.getEntry(path)
	if err != nil {
		return nil, err
	}

	if e.Filemode == git2go.FilemodeLink {
		// dereference symlink
		b, err := fs.repo.LookupBlob(e.Id)
		if err != nil {
			return nil, err
		}

		derefPath := string(b.Contents())
		fi, err := fs.Lstat(derefPath)
		if err != nil {
			return nil, err
		}

		// Use original filename.
		fi.(*util.FileInfo).Name_ = filepath.Base(path)

		return fi, nil
	}

	fi, err := fs.makeFileInfo(path, e)
	if err != nil {
		return nil, err
	}
	fi.ModTime_ = mtime

	return fi, nil
}

func (fs *gitFSLibGit2) getModTime() (time.Time, error) {
	commit, err := fs.repo.LookupCommit(fs.oid)
	if err != nil {
		return time.Time{}, err
	}
	return commit.Author().When, nil
}

func (fs *gitFSLibGit2) makeFileInfo(path string, e *git2go.TreeEntry) (*util.FileInfo, error) {
	switch e.Type {
	case git2go.ObjectBlob:
		return fs.fileInfo(e)
	case git2go.ObjectTree:
		return fs.dirInfo(e), nil
	case git2go.ObjectCommit:
		submod, err := fs.repo.Submodules.Lookup(path)
		if err != nil {
			return nil, err
		}

		// TODO(sqs): add (*Submodule).Free to git2go and defer submod.Free()
		// below when that method has been added.
		//
		// defer submod.Free()

		return &util.FileInfo{
			Name_: e.Name,
			Mode_: vcs.ModeSubmodule,
			Sys_: vcs.SubmoduleInfo{
				URL:      submod.Url(),
				CommitID: vcs.CommitID(e.Id.String()),
			},
		}, nil
	}

	return nil, fmt.Errorf("unexpected object type %v while making file info (expected blob, tree, or commit)", e.Type)
}

func (fs *gitFSLibGit2) fileInfo(e *git2go.TreeEntry) (*util.FileInfo, error) {
	b, err := fs.repo.LookupBlob(e.Id)
	if err != nil {
		return nil, err
	}
	defer b.Free()

	var sys interface{}
	var mode os.FileMode
	if e.Filemode == git2go.FilemodeBlobExecutable {
		mode |= 0111
	}
	if e.Filemode == git2go.FilemodeLink {
		mode |= os.ModeSymlink

		// Dereference symlink.
		b, err := fs.repo.LookupBlob(e.Id)
		if err != nil {
			return nil, err
		}
		defer b.Free()
		sys = vcs.SymlinkInfo{Dest: string(b.Contents())}
	}

	return &util.FileInfo{
		Name_: e.Name,
		Size_: b.Size(),
		Mode_: mode,
		Sys_:  sys,
	}, nil
}

func (fs *gitFSLibGit2) dirInfo(e *git2go.TreeEntry) *util.FileInfo {
	return &util.FileInfo{
		Name_: e.Name,
		Mode_: os.ModeDir,
	}
}

func (fs *gitFSLibGit2) ReadDir(path string) ([]os.FileInfo, error) {
	fs.repoEditLock.RLock()
	defer fs.repoEditLock.RUnlock()

	path = filepath.Clean(internal.Rel(path))

	var subtree *git2go.Tree
	if path == "." {
		subtree = fs.tree
	} else {
		e, err := fs.getEntry(path)
		if err != nil {
			return nil, err
		}

		subtree, err = fs.repo.LookupTree(e.Id)
		if err != nil {
			return nil, err
		}
	}

	fis := make([]os.FileInfo, int(subtree.EntryCount()))
	for i := uint64(0); i < subtree.EntryCount(); i++ {
		e := subtree.EntryByIndex(i)
		fi, err := fs.makeFileInfo(filepath.Join(path, e.Name), e)
		if err != nil {
			return nil, err
		}
		fis[i] = fi
	}

	return fis, nil
}

func (fs *gitFSLibGit2) String() string {
	return fmt.Sprintf("git repository %s commit %s (libgit2)", fs.dir, fs.at)
}

func isErrIterOver(err error) bool {
	if e, ok := err.(*git2go.GitError); ok {
		return e != nil && e.Code == git2go.ErrIterOver
	}
	return false
}

func standardizeLibGit2Error(err error) error {
	if err != nil && strings.Contains(err.Error(), "does not exist in the given tree") {
		return os.ErrNotExist
	}
	return err
}
