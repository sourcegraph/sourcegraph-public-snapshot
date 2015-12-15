package git

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/tools/godoc/vfs"

	"github.com/shazow/go-git"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/internal"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/util"
)

type filesystem struct {
	dir  string
	oid  string
	tree *git.Tree

	repo *git.Repository
}

func (fs *filesystem) readFileBytes(name string) ([]byte, error) {
	e, err := fs.tree.GetTreeEntryByPath(name)
	if err != nil {
		return nil, err
	}

	switch e.Type {
	case git.ObjectBlob:
		reader, err := e.Blob().Data()
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		return b, nil
	case git.ObjectCommit:
		// Return empty for a submodule for now.
		return nil, nil
	}
	return nil, fmt.Errorf("read unexpected entry type %q (expected blob or submodule(commit))", e.Type)
}

func (fs *filesystem) Open(name string) (vfs.ReadSeekCloser, error) {
	name = internal.Rel(name)

	b, err := fs.readFileBytes(name)
	if err != nil {
		return nil, err
	}
	return util.NopCloser{ReadSeeker: bytes.NewReader(b)}, nil
}

func (fs *filesystem) Lstat(path string) (os.FileInfo, error) {
	path = filepath.Clean(internal.Rel(path))

	mtime, err := fs.getModTime()
	if err != nil {
		return nil, err
	}

	if path == "." {
		return &util.FileInfo{Mode_: os.ModeDir, ModTime_: mtime}, nil
	}

	e, err := fs.tree.GetTreeEntryByPath(path)
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

func (fs *filesystem) Stat(path string) (os.FileInfo, error) {
	path = filepath.Clean(internal.Rel(path))

	mtime, err := fs.getModTime()
	if err != nil {
		return nil, err
	}

	if path == "." {
		return &util.FileInfo{Mode_: os.ModeDir, ModTime_: mtime}, nil
	}

	e, err := fs.tree.GetTreeEntryByPath(path)
	if err != nil {
		return nil, err
	}

	if e.EntryMode() == git.ModeSymlink {
		// Dereference symlink.
		reader, err := e.Blob().Data()
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		fi, err := fs.Lstat(string(b))
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

func (fs *filesystem) getModTime() (time.Time, error) {
	commit, err := fs.repo.GetCommit(fs.oid)
	if err != nil {
		return time.Time{}, err
	}
	return commit.Author.When, nil
}

func (fs *filesystem) makeFileInfo(path string, e *git.TreeEntry) (*util.FileInfo, error) {
	switch e.Type {
	case git.ObjectBlob:
		return fs.fileInfo(e)
	case git.ObjectTree:
		return fs.dirInfo(e)
	case git.ObjectCommit:
		return fs.submoduleInfo(path, e)
	}

	return nil, fmt.Errorf("unexpected object type %v while making file info (expected blob, tree, or commit)", e.Type)
}

func (fs *filesystem) submoduleInfo(path string, e *git.TreeEntry) (*util.FileInfo, error) {
	// TODO: Cache submodules?
	subs, err := e.Tree().GetSubmodules()
	if err != nil {
		return nil, err
	}
	var found *git.Submodule
	for _, sub := range subs {
		if sub.Path == path {
			found = sub
			break
		}
	}

	if found == nil {
		return nil, fmt.Errorf("submodule not found: %s", path)
	}

	return &util.FileInfo{
		Name_: e.Name(),
		Mode_: vcs.ModeSubmodule,
		Sys_: vcs.SubmoduleInfo{
			URL:      found.URL,
			CommitID: vcs.CommitID(e.Id.String()),
		},
	}, nil
}

func (fs *filesystem) fileInfo(e *git.TreeEntry) (*util.FileInfo, error) {
	var sys interface{}
	var mode os.FileMode
	if e.EntryMode() == git.ModeExec {
		mode |= 0111
	}
	if e.EntryMode() == git.ModeSymlink {
		mode |= os.ModeSymlink

		// Dereference symlink.
		reader, err := e.Blob().Data()
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}

		sys = vcs.SymlinkInfo{Dest: string(b)}
	}

	return &util.FileInfo{
		Name_: e.Name(),
		Size_: e.Size(),
		Mode_: mode,
		Sys_:  sys,
	}, nil
}

func (fs *filesystem) dirInfo(e *git.TreeEntry) (*util.FileInfo, error) {
	return &util.FileInfo{
		Name_: e.Name(),
		Mode_: os.ModeDir,
	}, nil
}

func (fs *filesystem) ReadDir(path string) ([]os.FileInfo, error) {
	path = filepath.Clean(internal.Rel(path))

	var subtree *git.Tree
	if path == "." {
		subtree = fs.tree
	} else {
		e, err := fs.tree.GetTreeEntryByPath(path)
		if err != nil {
			return nil, err
		}

		// FIXME: This looks redundant?
		subtree, err = fs.repo.GetTree(e.Id.String())
		if err != nil {
			return nil, err
		}
	}

	entries, err := subtree.ListEntries()
	if err != nil {
		return nil, err
	}

	fis := make([]os.FileInfo, 0, len(entries))
	for _, e := range entries {
		fi, err := fs.makeFileInfo(filepath.Join(path, e.Name()), e)
		if err != nil {
			return nil, err
		}
		fis = append(fis, fi)
	}

	return fis, nil
}

func (fs *filesystem) String() string {
	return fmt.Sprintf("git repository %s commit %s (gogit)", fs.dir, fs.oid)
}
