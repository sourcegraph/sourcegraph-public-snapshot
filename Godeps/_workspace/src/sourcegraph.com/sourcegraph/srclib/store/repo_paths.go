package store

import (
	"strings"

	"github.com/kr/fs"
	"sourcegraph.com/sourcegraph/rwvfs"
)

// RepoPaths specifies how to generate and list repos in a multi-repo
// store.
type RepoPaths interface {
	// RepoToPath takes a repo identifier (URI) and returns the path
	// components of the path under which its data should be stored in
	// the multi-repo store.
	//
	// The path components are joined using the VFS's Join method to
	// construct the full subpath. Using the VFS's Join method ensures
	// that the RepoToPath func uses the filesystem's path separator
	// (which usually but not always '/').
	RepoToPath(string) []string

	// PathToRepo is the inverse of RepoToPath.
	PathToRepo(path []string) string

	// ListRepoPaths returns a lexicographically sorted list of repo
	// subpaths (originally created using the RepoToPath func). Only
	// those that sort lexicographically after the "after" arg are
	// returned. If "after" is empty, all keys are returned (up to the
	// max).
	ListRepoPaths(vfs rwvfs.WalkableFileSystem, after string, max int) ([][]string, error)
}

// DefaultRepoPaths is the default repo path configuration for
// FS-backed multi-repo stores. It stores repos underneath the
// "${REPO}/.srclib-store" dir.
var DefaultRepoPaths defaultRepoPaths

type defaultRepoPaths struct{}

// RepoToPath implements RepoPaths.
func (defaultRepoPaths) RepoToPath(repo string) []string {
	p := strings.Split(repo, "/")
	p = append(p, SrclibStoreDir)
	return p
}

// PathToRepo implements RepoPaths.
func (defaultRepoPaths) PathToRepo(path []string) string {
	return strings.Join(path[:len(path)-1], "/")
}

// ListRepoPaths implements RepoPaths.
func (defaultRepoPaths) ListRepoPaths(vfs rwvfs.WalkableFileSystem, after string, max int) ([][]string, error) {
	var paths [][]string
	w := fs.WalkFS(".", rwvfs.Walkable(vfs))
	for w.Step() {
		if err := w.Err(); err != nil {
			return nil, err
		}
		fi := w.Stat()
		if w.Path() >= after && fi.Mode().IsDir() {
			if fi.Name() == SrclibStoreDir {
				w.SkipDir()
				// NOTE: This assumes that the vfs's path
				// separator is "/", which is not true in general.
				paths = append(paths, strings.Split(w.Path(), "/"))
				if max != 0 && len(paths) >= max {
					break
				}
				continue
			}
			if fi.Name() != "." && strings.HasPrefix(fi.Name(), ".") {
				w.SkipDir()
				continue
			}
		}
	}
	return paths, nil
}
