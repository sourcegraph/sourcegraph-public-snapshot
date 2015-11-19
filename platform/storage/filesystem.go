package storage

import (
	"fmt"
	"os"

	"src.sourcegraph.com/vfs"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// fileSystem implements the vfs.FileSystem interface on top of the gRPC API.
type fileSystem struct {
	ctx     context.Context
	client  *sourcegraph.Client
	appName string
	repo    string
}

// storageName is a utility function which returns a new StorageName given it's
// name. The AppName and Repo fields are handled for you.
func (fs *fileSystem) storageName(name string) *sourcegraph.StorageName {
	return &sourcegraph.StorageName{
		AppName: fs.appName,
		Repo:    fs.repo,
		Name:    name,
	}
}

// String implements the fmt.Stringer interface.
func (fs *fileSystem) String() string {
	return fmt.Sprintf("FileSystem(AppName=%q, Repo=%q)", fs.appName, fs.repo)
}

// Create implements the vfs.FileSystem interface.
func (fs *fileSystem) Create(name string) (vfs.File, error) {
	storageName := fs.storageName(name)
	ioErr, grpcErr := fs.client.Storage.Create(fs.ctx, storageName)
	if grpcErr != nil {
		return nil, grpcErr
	}
	if isStorageError(ioErr) {
		return nil, storageError(ioErr)
	}
	return &file{
		fs:   fs,
		name: storageName,
	}, nil
}

// RemoveAll implements the vfs.FileSystem interface.
func (fs *fileSystem) RemoveAll(name string) error {
	ioErr, grpcErr := fs.client.Storage.RemoveAll(fs.ctx, fs.storageName(name))
	if grpcErr != nil {
		return grpcErr
	}
	return storageError(ioErr)
}

// Open implements the vfs.FileSystem interface.
func (fs *fileSystem) Open(name string) (vfs.File, error) {
	_, err := fs.Stat(name)
	if err != nil {
		return nil, err
	}
	return &file{
		fs:   fs,
		name: fs.storageName(name),
	}, nil
}

// Stat implements the vfs.FileSystem interface.
func (fs *fileSystem) Stat(path string) (os.FileInfo, error) {
	resp, grpcErr := fs.client.Storage.Stat(fs.ctx, fs.storageName(path))
	if grpcErr != nil {
		return nil, grpcErr
	}
	if isStorageError(resp.Error) {
		return nil, storageError(resp.Error)
	}
	return fileInfo{resp.Info}, nil
}

// ReadDir implements the vfs.FileSystem interface.
func (fs *fileSystem) ReadDir(path string) ([]os.FileInfo, error) {
	resp, grpcErr := fs.client.Storage.ReadDir(fs.ctx, fs.storageName(path))
	if grpcErr != nil {
		return nil, grpcErr
	}
	if isStorageError(resp.Error) {
		return nil, storageError(resp.Error)
	}
	var infos []os.FileInfo
	for _, fi := range resp.Info {
		infos = append(infos, fileInfo{fi})
	}
	return infos, nil
}
