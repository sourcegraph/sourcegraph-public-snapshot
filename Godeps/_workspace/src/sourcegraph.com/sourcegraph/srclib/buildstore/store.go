package buildstore

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"code.google.com/p/rog-go/parallel"

	"github.com/kr/fs"

	"sourcegraph.com/sourcegraph/rwvfs"
)

// BuildDataDirName is the name of the directory in which local
// repository build data is stored, relative to the top-level dir of a
// VCS repository.
var BuildDataDirName = ".srclib-cache"

// A MultiStore contains RepoBuildStores for multiple repositories.
type MultiStore struct {
	fs rwvfs.WalkableFileSystem
}

// NewMulti creates a new multi-repo build store.
func NewMulti(fs rwvfs.FileSystem) *MultiStore {
	return &MultiStore{rwvfs.Walkable(fs)}
}

func (s *MultiStore) RepoBuildStore(repoURI string) (RepoBuildStore, error) {
	path := filepath.Clean(string(repoURI))
	return Repo(rwvfs.Walkable(rwvfs.Sub(s.fs, path))), nil
}

// A RepoBuildStore stores and exposes a repository's build data in a
// VFS.
type RepoBuildStore interface {
	// Commit returns a VFS for accessing and writing build data for a
	// specific commit.
	Commit(commitID string) rwvfs.WalkableFileSystem

	// FilePath returns the path (from the repo build store's root) to
	// a file at the specified commit ID.
	FilePath(commitID string, file string) string
}

// Repo creates a new single-repository build store rooted at the
// given filesystem.
func Repo(repoStoreFS rwvfs.WalkableFileSystem) RepoBuildStore {
	return &repoBuildStore{repoStoreFS}
}

// LocalRepo creates a new single-repository build store for the VCS
// repository whose top-level directory is repoDir.
//
// The store is laid out as follows:
//
//   .                the root dir of repoStoreFS
//   <COMMITID>/**/*  build data for a specific commit
func LocalRepo(repoDir string) (RepoBuildStore, error) {
	storeDir := filepath.Join(repoDir, BuildDataDirName)
	if err := os.Mkdir(storeDir, 0700); err != nil && !os.IsExist(err) {
		return nil, err
	}
	fs := rwvfs.OS(storeDir)
	setCreateParentDirs(fs)
	return Repo(rwvfs.Walkable(fs)), nil
}

func setCreateParentDirs(fs rwvfs.FileSystem) {
	type createParents interface {
		CreateParentDirs(bool)
	}
	if fs, ok := fs.(createParents); ok {
		fs.CreateParentDirs(true)
	}
}

type repoBuildStore struct {
	fs rwvfs.WalkableFileSystem
}

func (s *repoBuildStore) Commit(commitID string) rwvfs.WalkableFileSystem {
	path := s.commitPath(commitID)

	// Dereference path if path refers to a symlink, so that we can
	// walk the tree.

	e, _ := s.fs.Lstat(path)
	if e != nil && e.Mode()&os.ModeSymlink != 0 {
		if fs, ok := s.fs.(rwvfs.LinkFS); ok {
			var err error
			dst, err := fs.ReadLink(path)
			if err == nil {
				path = dst
			} else if err == rwvfs.ErrOutsideRoot && FollowCrossFSSymlinks {
				return rwvfs.Walkable(rwvfs.OS(dst))
			} else {
				log.Printf("Failed to read symlink %s: %s. Using non-dereferenced path.", path, err)
			}
		} else {
			log.Printf("Repository build store path for commit %s is a symlink, but the current VFS %s doesn't support dereferencing symlinks.", commitID, s.fs)
		}
	}
	return rwvfs.Walkable(rwvfs.Sub(s.fs, path))
}

func (s *repoBuildStore) commitPath(commitID string) string { return commitID }

func (s *repoBuildStore) FilePath(commitID, path string) string {
	return filepath.Join(s.commitPath(commitID), path)
}

// RemoveAllDataForCommit removes all files and directories from the
// repo build store for the given commit.
func RemoveAllDataForCommit(s RepoBuildStore, commitID string) error {
	commitFS := s.Commit(commitID)
	return RemoveAll(".", commitFS)
}

// RemoveAll removes a tree recursively.
func RemoveAll(path string, vfs rwvfs.WalkableFileSystem) error {
	w := fs.WalkFS(path, vfs)

	remove := func(par *parallel.Run, path string) {
		par.Do(func() error { return vfs.Remove(path) })
	}

	var dirs []string // remove dirs after removing all files
	filesPar := parallel.NewRun(20)
	for w.Step() {
		if err := w.Err(); err != nil {
			return err
		}
		if w.Stat().IsDir() {
			dirs = append(dirs, w.Path())
		} else {
			remove(filesPar, w.Path())
		}
	}

	if err := filesPar.Wait(); err != nil {
		return err
	}

	dirsPar := parallel.NewRun(20)
	sort.Sort(sort.Reverse(sort.StringSlice(dirs))) // reverse so we delete leaf dirs first
	for _, dir := range dirs {
		remove(dirsPar, dir)
	}
	return dirsPar.Wait()
}

func BuildDataExistsForCommit(s RepoBuildStore, commitID string) (bool, error) {
	cfs := s.Commit(commitID)
	_, err := cfs.Stat(".")
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// FollowCrossFSSymlinks is whether symlinks inside a buildstore
// should be followed if they cross filesystem boundaries. This should
// only be true during testing. Setting it to true during normal
// operation could make it possible for an attacker who uploads
// symlinks to a build data store to read files on your local
// filesystem.
var FollowCrossFSSymlinks, _ = strconv.ParseBool(os.Getenv("SRCLIB_FOLLOW_CROSS_FS_SYMLINKS"))
