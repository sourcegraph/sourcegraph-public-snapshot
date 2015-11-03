package storage

import (
	"errors"
	"fmt"
	"os"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// file implements the vfs.File interface on top of the gRPC API.
type file struct {
	fs     *fileSystem
	name   *sourcegraph.StorageName
	offset int64
}

// Name implements the vfs.File interface.
func (f *file) Name() string {
	return f.name.Name
}

// Truncate implements the vfs.File interface.
func (f *file) Truncate(size int64) error {
	// TODO(slimsag): implement Truncate
	panic("Truncate is not implement via gRPC Storage API")
}

// String implements the fmt.Stringer interface.
func (f *file) String() string {
	return fmt.Sprintf("File(%q, FileSystem=%v)", f.name.Name, f.fs)
}

// Read implements the io.Reader interface.
func (f *file) Read(p []byte) (n int, err error) {
	resp, grpcErr := f.fs.client.Storage.Read(f.fs.ctx, &sourcegraph.StorageReadOp{
		Name:   *f.name,
		Offset: f.offset,
		Count:  int64(len(p)),
	})
	if grpcErr != nil {
		return 0, grpcErr
	}
	f.offset += int64(len(resp.Data))
	copy(p, resp.Data)
	return len(resp.Data), storageError(resp.Error)
}

// Write implements the io.Writer interface.
func (f *file) Write(p []byte) (n int, err error) {
	resp, grpcErr := f.fs.client.Storage.Write(f.fs.ctx, &sourcegraph.StorageWriteOp{
		Name:   *f.name,
		Offset: f.offset,
		Data:   p,
	})
	if grpcErr != nil {
		return 0, grpcErr
	}
	f.offset += resp.Wrote
	return int(resp.Wrote), storageError(resp.Error)
}

// Seek implements the io.Seeker interface.
func (f *file) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case os.SEEK_SET:
		f.offset = offset
	case os.SEEK_CUR:
		f.offset += offset
	case os.SEEK_END:
		fi, err := f.fs.Lstat(f.name.Name)
		if err != nil {
			return 0, err
		}
		f.offset = fi.Size() - offset
	default:
		panic("File.Seek: invalid whence value")
	}
	if f.offset < 0 {
		return nil, errors.New("seek to negative offset")
	}
	return f.offset, nil
}

// Close implements the io.Closer interface.
func (f *file) Close() error {
	ioErr, grpcErr := f.fs.client.Storage.Close(f.fs.ctx, f.name)
	if grpcErr != nil {
		return grpcErr
	}
	return storageError(ioErr)
}
