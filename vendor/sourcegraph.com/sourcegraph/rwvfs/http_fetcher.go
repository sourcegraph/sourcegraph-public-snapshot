package rwvfs

import (
	"errors"
	"fmt"
	"log"
	"os"

	"golang.org/x/tools/godoc/vfs"
)

// A RangeOpener is a filesystem that can efficiently open a range
// within a file (instead of the whole file).
type RangeOpener interface {
	// OpenRange returns a vfs.ReadSeekCloser to read the specified
	// range from the named file. The rangeHeader parameter must be in
	// the format of HTTP Range headers (e.g., "bytes=123-456"). An
	// empty rangeHeader is interpreted to mean the entire file.
	OpenRange(name string, rangeHeader string) (vfs.ReadSeekCloser, error)
}

// OpenFetcher creates a new vfs.ReadSeekCloser based on the named
// file in fs that optimistically buffers reads. It is intended to be
// used when fs is a network filesystem with relatively high RTT
// (10msec+).
func OpenFetcher(fs RangeOpener, name string) (vfs.ReadSeekCloser, error) {
	return &explicitFetchFile{name: name, fs: fs, autofetch: true}, nil
}

type explicitFetchFile struct {
	name               string
	fs                 RangeOpener
	startByte, endByte int64
	rc                 vfs.ReadSeekCloser
	autofetch          bool
}

var vlog = log.New(os.Stderr, "rwvfs: ", 0)

func (f *explicitFetchFile) Read(p []byte) (n int, err error) {
	ofs, err := f.Seek(0, 1) // get current offset
	if err != nil {
		return 0, err
	}
	if start, end := ofs, ofs+int64(len(p)); !f.isFetched(start, end) {
		if !f.autofetch {
			return 0, fmt.Errorf("rwvfs: range %d-%d not fetched (%d-%d fetched; offset %d)", start, end, f.startByte, f.endByte, ofs)
		}
		const x = 4 // overfetch factor (because network RTT >> network throughput)
		fetchEnd := end + (end-start)*x
		vlog.Printf("Autofetching range %d-%d because read of unfetched %d-%d attempted (%d bytes)", start, fetchEnd, start, end, len(p))
		if err := f.Fetch(start, fetchEnd); err != nil {
			return 0, err
		}
	}
	return f.rc.Read(p)
}

func (f *explicitFetchFile) isFetched(start, end int64) bool {
	return f.rc != nil && start <= end && start >= f.startByte && end <= f.endByte
}

func (f *explicitFetchFile) Fetch(start, end int64) error {
	if f.isFetched(start, end) {
		// Already prefetched.
		vlog.Printf("Already fetched %d-%d (fetched range is %d-%d)", start, end, f.startByte, f.endByte)
		return nil
	}

	// Close existing open reader (if any).
	if err := f.Close(); err != nil {
		return err
	}

	rng := fmt.Sprintf("bytes=%d-%d", start, end)
	var err error
	f.rc, err = f.fs.OpenRange(f.name, rng)
	if err == nil {
		f.startByte = start
		f.endByte = end
	}
	return err
}

var errRelOfs = errors.New("rwvfs: seek to offset relative to end of file is not supported")

func (f *explicitFetchFile) Seek(offset int64, whence int) (int64, error) {
	if f.rc == nil {
		return 0, errors.New("rwvfs: must call Fetch before Seek")
	}

	switch whence {
	case 0:
		offset -= f.startByte
	case 2:
		return 0, errRelOfs
	}
	n, err := f.rc.Seek(offset, whence)
	n += f.startByte
	return n, err
}

func (f *explicitFetchFile) Close() error {
	if f.rc != nil {
		err := f.rc.Close()
		f.rc = nil
		f.startByte = 0
		f.endByte = 0
		return err
	}
	return nil
}

// SetAutofetch sets whether data is automatically fetched when a read
// requires it. It is used during tests only.
func (f *explicitFetchFile) SetAutofetch(autofetch bool) { f.autofetch = autofetch }
