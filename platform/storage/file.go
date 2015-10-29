package storage

import (
	"fmt"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// file implements the File interface on top of the gRPC API.
type file struct {
	fs     *fileSystem
	name   *sourcegraph.StorageName
	offset int64
}

// Name implements the File interface.
func (f *file) Name() string {
	return f.name.Name
}

// Truncate implements the File interface.
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
	copy(resp.Data, p)
	return len(resp.Data), storageError(&resp.Error)
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
	return int(resp.Wrote), storageError(resp.Error)
}

// Seek implements the io.Seeker interface.
func (f *file) Seek(offset int64, whence int) (int64, error) {
	if offset < 0 {
		panic("File.Seek: cannot seek to a negative offset")
	}
	switch whence {
	case 0:
		f.offset = offset
	case 1:
		f.offset += offset
	case 2:
		fi, err := f.fs.Lstat(f.name.Name)
		if err != nil {
			return 0, err
		}
		f.offset = fi.Size() - offset
	default:
		panic("File.Seek: invalid whence value")
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
