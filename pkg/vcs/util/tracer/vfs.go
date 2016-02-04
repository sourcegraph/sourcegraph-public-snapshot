package tracer

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/appdash"
)

// readSeekCloser implements the vfs.ReadSeekCloser interface and emits a single
// event upon close (lightly representitive of time it took to read the file).
type readSeekCloser struct {
	vfs.ReadSeekCloser
	start time.Time
	name  string
	rec   *appdash.Recorder
}

// Close implements the io.Closer interface.
func (r readSeekCloser) Close() error {
	err := r.ReadSeekCloser.Close()
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.Repository.FileSystem.Open -> Close",
		Args:      r.name,
		StartTime: r.start,
		EndTime:   time.Now(),
	})
	return err
}

// fileSystem implements the vfs.FileSystem interface.
type fileSystem struct {
	fs  vfs.FileSystem
	rec *appdash.Recorder
}

// Open implements the vfs.Opener interface.
func (fs fileSystem) Open(name string) (vfs.ReadSeekCloser, error) {
	start := time.Now()
	f, err := fs.fs.Open(name)
	end := time.Now()
	fs.rec.Child().Event(GoVCS{
		Name:      "vcs.Repository.FileSystem.Open",
		Args:      name,
		StartTime: start,
		EndTime:   end,
	})
	if err != nil {
		return nil, err
	}
	return readSeekCloser{
		ReadSeekCloser: f,
		start:          end,
		name:           name,
		rec:            fs.rec,
	}, nil
}

// Lstat implements the vfs.FileSystem interface.
func (fs fileSystem) Lstat(path string) (os.FileInfo, error) {
	start := time.Now()
	fi, err := fs.fs.Lstat(path)
	fs.rec.Child().Event(GoVCS{
		Name:      "vcs.Repository.FileSystem.Lstat",
		Args:      fmt.Sprintf("%#v", path),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return fi, err
}

// Stat implements the vfs.FileSystem interface.
func (fs fileSystem) Stat(path string) (os.FileInfo, error) {
	start := time.Now()
	fi, err := fs.fs.Stat(path)
	fs.rec.Child().Event(GoVCS{
		Name:      "vcs.Repository.FileSystem.Stat",
		Args:      fmt.Sprintf("%#v", path),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return fi, err
}

// ReadDir implements the vfs.FileSystem interface.
func (fs fileSystem) ReadDir(path string) ([]os.FileInfo, error) {
	start := time.Now()
	fis, err := fs.fs.ReadDir(path)
	fs.rec.Child().Event(GoVCS{
		Name:      "vcs.Repository.FileSystem.ReadDir",
		Args:      fmt.Sprintf("%#v", path),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return fis, err
}

// String implements the vfs.FileSystem interface.
func (fs fileSystem) String() string {
	return fs.fs.String()
}
