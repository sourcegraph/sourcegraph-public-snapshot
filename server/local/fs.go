package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"sort"
	"sync"

	"src.sourcegraph.com/sourcegraph/pkg/vcsclient"

	"github.com/sqs/fileset"

	"sourcegraph.com/sqs/pbtypes"

	"golang.org/x/tools/godoc/vfs"
)

// getFileWithOptions gets a file and observes the options specified in opt.
func getFileWithOptions(fs vfs.FileSystem, path string, opt vcsclient.GetFileOptions) (*vcsclient.FileWithRange, error) {
	fi, err := fs.Lstat(path)
	if err != nil {
		return nil, err
	}

	e := newTreeEntry(fi)
	fwr := vcsclient.FileWithRange{TreeEntry: e}

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

		if empty := (vcsclient.GetFileOptions{}); opt != empty {
			fr, _, err := computeFileRange(contents, opt)
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
func readDir(fs vfs.FileSystem, base string, recurseSingleSubfolderLimit int, first bool) ([]*vcsclient.TreeEntry, error) {
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
		te         = make([]*vcsclient.TreeEntry, len(entries))
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

func newTreeEntry(fi os.FileInfo) *vcsclient.TreeEntry {
	e := &vcsclient.TreeEntry{
		Name:    fi.Name(),
		Size_:   fi.Size(),
		ModTime: pbtypes.NewTimestamp(fi.ModTime()),
	}
	if fi.Mode().IsDir() {
		e.Type = vcsclient.DirEntry
	} else if fi.Mode().IsRegular() {
		e.Type = vcsclient.FileEntry
	} else if fi.Mode()&os.ModeSymlink != 0 {
		e.Type = vcsclient.SymlinkEntry
	}
	return e
}

// computeFileRange determines the actual file range according to the
// input range parameter. For example, if input has a line range set,
// the returned FileRange will contain the byte range that corresponds
// to the input line range.
func computeFileRange(data []byte, opt vcsclient.GetFileOptions) (*vcsclient.FileRange, *fileset.File, error) {
	fr := opt.FileRange // alias for brevity

	fset := fileset.NewFileSet()
	f := fset.AddFile("", 1, len(data))
	f.SetLinesForContent(data)

	if opt.EntireFile || (fr.StartLine == 0 && fr.EndLine == 0 && fr.StartByte == 0 && fr.EndByte == 0) {
		fr.StartLine, fr.EndLine = 0, 0
		fr.StartByte, fr.EndByte = 0, int64(len(data))
	}

	lines := fr.StartLine != 0 || fr.EndLine != 0
	bytes := fr.StartByte != 0 || fr.EndByte != 0
	if lines && bytes {
		return nil, nil, fmt.Errorf("must specify a line range OR a byte range, not both (%+v)", fr)
	}

	// TODO(sqs): fix up the sketchy int conversions

	if lines {
		// Given line range, validate it and return byte range.
		if fr.StartLine == 0 {
			fr.StartLine = 1 // 1-indexed
		}
		if fr.StartLine == 1 && fr.EndLine == 1 && f.LineCount() == 0 {
			// Empty file.
			return &fr, f, nil
		}
		if fr.StartLine < 0 || fr.StartLine > int64(f.LineCount()) {
			return nil, nil, fmt.Errorf("start line %d out of bounds (%d lines total)", fr.StartLine, f.LineCount())
		}
		if fr.EndLine < 0 {
			return nil, nil, fmt.Errorf("end line %d out of bounds (%d lines total)", fr.EndLine, f.LineCount())
		}
		if fr.StartLine > fr.EndLine {
			return nil, nil, fmt.Errorf("start line (%d) cannot be greater than end line (%d) (%d lines total)", fr.StartLine, fr.EndLine, f.LineCount())
		}

		if count := int64(f.LineCount()); fr.EndLine > count || fr.EndLine == 0 {
			fr.EndLine = count
		}
		fr.StartByte, fr.EndByte = int64(f.LineOffset(int(fr.StartLine))), int64(f.LineEndOffset(int(fr.EndLine)))
	} else if bytes {
		if fr.StartByte < 0 || fr.StartByte > int64(len(data)-1) {
			return nil, nil, fmt.Errorf("start byte %d out of bounds (%d bytes total)", fr.StartByte, len(data))
		}
		if fr.EndByte < 0 || fr.EndByte > int64(len(data)) {
			return nil, nil, fmt.Errorf("end byte %d out of bounds (%d bytes total)", fr.EndByte, len(data))
		}
		if fr.StartByte > fr.EndByte {
			return nil, nil, fmt.Errorf("start byte (%d) cannot be greater than end byte (%d) (%d bytes total)", fr.StartByte, fr.EndByte, len(data))
		}

		fr.StartLine, fr.EndLine = int64(f.Line(f.Pos(int(fr.StartByte)))), int64(f.Line(f.Pos(int(fr.EndByte))))
		if opt.ExpandContextLines > 0 {
			fr.StartLine -= int64(opt.ExpandContextLines)
			if fr.StartLine < 1 {
				fr.StartLine = 1
			}
			fr.EndLine += int64(opt.ExpandContextLines)
			if fr.EndLine > int64(f.LineCount()) {
				fr.EndLine = int64(f.LineCount())
			}
		}
		if opt.ExpandContextLines > 0 || opt.FullLines {
			fr.StartByte, fr.EndByte = int64(f.LineOffset(int(fr.StartLine))), int64(f.LineEndOffset(int(fr.EndLine)))
		}
	}

	return &fr, f, nil
}

type TreeEntriesByTypeByName []*vcsclient.TreeEntry

func (v TreeEntriesByTypeByName) Len() int      { return len(v) }
func (v TreeEntriesByTypeByName) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v TreeEntriesByTypeByName) Less(i, j int) bool {
	// Sort dirs before everything else.
	if v[i].Type == vcsclient.DirEntry && v[j].Type != vcsclient.DirEntry {
		return true
	}
	if v[i].Type != vcsclient.DirEntry && v[j].Type == vcsclient.DirEntry {
		return false
	}
	return v[i].Name < v[j].Name
}
