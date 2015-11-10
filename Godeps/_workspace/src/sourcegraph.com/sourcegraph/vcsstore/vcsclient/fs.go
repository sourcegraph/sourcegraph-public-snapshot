package vcsclient

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"

	"sort"

	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sqs/pbtypes"
)

type FileSystem interface {
	vfs.FileSystem
	Get(path string) (*TreeEntry, error)
}

type repositoryFS struct {
	at   vcs.CommitID
	repo *repository
}

var _ FileSystem = &repositoryFS{}

func (fs *repositoryFS) Open(name string) (vfs.ReadSeekCloser, error) {
	e, err := fs.Get(name)
	if err != nil {
		return nil, err
	}

	return nopCloser{bytes.NewReader(e.Contents)}, nil
}

func (fs *repositoryFS) Lstat(path string) (os.FileInfo, error) {
	e, err := fs.Get(path)
	if err != nil {
		return nil, err
	}

	return e.Stat()
}

func (fs *repositoryFS) Stat(path string) (os.FileInfo, error) {
	// TODO(sqs): follow symlinks (as Stat specification requires)
	e, err := fs.Get(path)
	if err != nil {
		return nil, err
	}

	return e.Stat()
}

func (fs *repositoryFS) ReadDir(path string) ([]os.FileInfo, error) {
	e, err := fs.Get(path)
	if err != nil {
		return nil, err
	}

	fis := make([]os.FileInfo, len(e.Entries))
	for i, e := range e.Entries {
		fis[i], err = e.Stat()
		if err != nil {
			return nil, err
		}
	}

	return fis, nil
}

func (fs *repositoryFS) String() string {
	return fmt.Sprintf("repository %s commit %s (client)", fs.repo.repoPath, fs.at)
}

// Get returns the whole TreeEntry struct for a tree entry.
func (fs *repositoryFS) Get(path string) (*TreeEntry, error) {
	url, err := fs.url(path, nil)
	if err != nil {
		return nil, err
	}

	req, err := fs.repo.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	var entry *TreeEntry
	_, err = fs.repo.client.Do(req, &entry)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// FileWithRange is returned by GetFileWithOptions and includes the
// returned file's TreeEntry as well as the actual range of lines and
// bytes returned (based on the GetFileOptions parameters). That is,
// if Start/EndLine are set in GetFileOptions, this struct's
// Start/EndByte will be set to the actual start and end bytes of
// those specified lines, and so on for the other fields in
// GetFileOptions.
type FileWithRange struct {
	*TreeEntry
	FileRange // range of actual returned tree entry contents within file
}

// GetFileWithOptions gets a file and allows additional configuration
// of the range to return, etc.
func (fs *repositoryFS) GetFileWithOptions(path string, opt GetFileOptions) (*FileWithRange, error) {
	url, err := fs.url(path, opt)
	if err != nil {
		return nil, err
	}

	req, err := fs.repo.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	var file *FileWithRange
	_, err = fs.repo.client.Do(req, &file)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// A FileGetter is a repository FileSystem that can get files with
// extended range options (GetFileWithOptions).
//
// It's generally more efficient to use the client's implementation of
// the GetFileWithOptions method instead of calling the
// vcsclient.GetFileWithOptions func because the former causes only
// the requested range to be sent over the network, while the latter
// requests the whole file and narrows the range on the client side.
type FileGetter interface {
	GetFileWithOptions(path string, opt GetFileOptions) (*FileWithRange, error)
}

// GetFileWithOptions gets a file and observes the options specified
// in opt. If fs implements FileGetter, fs.GetFileWithOptions is
// called; otherwise the options are applied on the client side after
// fetching the whole file.
func GetFileWithOptions(fs vfs.FileSystem, path string, opt GetFileOptions) (*FileWithRange, error) {
	if fg, ok := fs.(FileGetter); ok {
		return fg.GetFileWithOptions(path, opt)
	}

	fi, err := fs.Lstat(path)
	if err != nil {
		return nil, err
	}

	e := newTreeEntry(fi)
	fwr := FileWithRange{TreeEntry: e}

	if fi.Mode().IsDir() {
		ee, err := readDir(fs, path, opt.RecurseSingleSubfolder, true)
		if err != nil {
			return nil, err
		}
		sort.Sort(TreeEntriesByTypeByName(ee))
		e.Entries = ee
	} else if fi.Mode().IsRegular() {
		f, err := fs.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		contents, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		e.Contents = contents

		if empty := (GetFileOptions{}); opt != empty {
			fr, _, err := ComputeFileRange(contents, opt)
			if err != nil {
				return nil, err
			}

			// Trim to only requested range.
			e.Contents = e.Contents[fr.StartByte:fr.EndByte]
			fwr.FileRange = *fr
		}
	}

	return &fwr, nil
}

// readDir uses the passed vfs.FileSystem to read from starting at the base path.
// If recurseSingleSubfolder is true, it will descend and include sub-folders
// with a single sub-folder inside. first should always be set to true, other values are used internally.
func readDir(fs vfs.FileSystem, base string, recurseSingleSubfolder bool, first bool) ([]*TreeEntry, error) {
	entries, err := fs.ReadDir(base)
	if err != nil {
		return nil, err
	}
	if recurseSingleSubfolder && !first && !singleSubDir(entries) {
		return nil, nil
	}
	te := make([]*TreeEntry, len(entries))
	for i, fi := range entries {
		te[i] = newTreeEntry(fi)
		if fi.Mode().IsDir() && recurseSingleSubfolder {
			ee, err := readDir(fs, path.Join(base, fi.Name()), recurseSingleSubfolder, false)
			if err != nil {
				return nil, err
			}
			te[i].Entries = ee
		}
	}
	return te, nil
}

func singleSubDir(entries []os.FileInfo) bool {
	return len(entries) == 1 && entries[0].IsDir()
}

func newTreeEntry(fi os.FileInfo) *TreeEntry {
	e := &TreeEntry{
		Name:    fi.Name(),
		Size:    fi.Size(),
		ModTime: pbtypes.NewTimestamp(fi.ModTime()),
	}
	if fi.Mode().IsDir() {
		e.Type = DirEntry
	} else if fi.Mode().IsRegular() {
		e.Type = FileEntry
	} else if fi.Mode()&os.ModeSymlink != 0 {
		e.Type = SymlinkEntry
	}
	return e
}

// url generates the URL to RouteRepoTreeEntry for the given path (all other
// route vars are taken from repositoryFS fields).
func (fs *repositoryFS) url(path string, opt interface{}) (*url.URL, error) {
	return fs.repo.url(RouteRepoTreeEntry, map[string]string{
		"CommitID": string(fs.at),
		"Path":     path,
	}, opt)
}

type nopCloser struct {
	io.ReadSeeker
}

func (nc nopCloser) Close() error { return nil }
