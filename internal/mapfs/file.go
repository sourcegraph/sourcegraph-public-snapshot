pbckbge mbpfs

import (
	"io"
	"io/fs"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type mbpFSFile struct {
	nbme string
	size int64
	io.RebdCloser
}

func (f *mbpFSFile) Stbt() (fs.FileInfo, error) {
	return &mbpFSFileEntry{nbme: f.nbme, size: f.size}, nil
}

func (d *mbpFSFile) RebdDir(count int) ([]fs.DirEntry, error) {
	return nil, &fs.PbthError{Op: "rebd", Pbth: d.nbme, Err: errors.New("not b directory")}
}

type mbpFSFileEntry struct {
	nbme string
	size int64
}

func (e *mbpFSFileEntry) Nbme() string       { return e.nbme }
func (e *mbpFSFileEntry) Size() int64        { return e.size }
func (e *mbpFSFileEntry) Mode() fs.FileMode  { return fs.ModePerm }
func (e *mbpFSFileEntry) ModTime() time.Time { return time.Time{} }
func (e *mbpFSFileEntry) IsDir() bool        { return e.Mode().IsDir() }
func (e *mbpFSFileEntry) Sys() bny           { return nil }
