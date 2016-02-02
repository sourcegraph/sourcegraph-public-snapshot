package vcsclient

import (
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"sort"
	"sync"

	"sourcegraph.com/sqs/pbtypes"

	"golang.org/x/tools/godoc/vfs"
)

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
		ee, err := readDir(fs, path, int(opt.RecurseSingleSubfolderLimit), true)
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
// If recurseSingleSubfolderLimit is non-zero, it will descend and include
// sub-folders with a single sub-folder inside. It will only inspect up to
// recurseSingleSubfolderLimit sub-folders. first should always be set to
// true, other values are used internally.
func readDir(fs vfs.FileSystem, base string, recurseSingleSubfolderLimit int, first bool) ([]*TreeEntry, error) {
	entries, err := fs.ReadDir(base)
	if err != nil {
		return nil, err
	}
	if recurseSingleSubfolderLimit > 0 && !first && !singleSubDir(entries) {
		return nil, nil
	}
	var (
		wg         sync.WaitGroup
		recurseErr error
		dirCount   = 0
		sem        = make(chan bool, runtime.GOMAXPROCS(0))
		te         = make([]*TreeEntry, len(entries))
	)
	for i, fi := range entries {
		te[i] = newTreeEntry(fi)
		if fi.Mode().IsDir() && dirCount < recurseSingleSubfolderLimit {
			dirCount++
			i, name := i, fi.Name()
			wg.Add(1)
			sem <- true
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				ee, err := readDir(fs, path.Join(base, name), recurseSingleSubfolderLimit, false)
				if err != nil {
					recurseErr = err
					return
				}
				te[i].Entries = ee
			}()
		}
	}
	wg.Wait()
	if recurseErr != nil {
		return nil, recurseErr
	}
	return te, nil
}

func singleSubDir(entries []os.FileInfo) bool {
	return len(entries) == 1 && entries[0].IsDir()
}

func newTreeEntry(fi os.FileInfo) *TreeEntry {
	e := &TreeEntry{
		Name:    fi.Name(),
		Size_:   fi.Size(),
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

type TreeEntriesByTypeByName []*TreeEntry

func (v TreeEntriesByTypeByName) Len() int      { return len(v) }
func (v TreeEntriesByTypeByName) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v TreeEntriesByTypeByName) Less(i, j int) bool {
	// Sort dirs before everything else.
	if v[i].Type == DirEntry && v[j].Type != DirEntry {
		return true
	}
	if v[i].Type != DirEntry && v[j].Type == DirEntry {
		return false
	}
	return v[i].Name < v[j].Name
}
