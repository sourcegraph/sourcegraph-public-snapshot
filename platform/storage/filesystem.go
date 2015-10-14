package storage

import (
	"fmt"
	"os"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

type fileSystem struct {
	ctx     context.Context
	client  *sourcegraph.Client
	appName string
	repo    *sourcegraph.RepoSpec
}

func (fs *fileSystem) storageName(name string) *sourcegraph.StorageName {
	return &sourcegraph.StorageName{
		AppName: fs.appName,
		Repo:    fs.repo,
		Name:    name,
	}
}

func (fs *fileSystem) String() string {
	return fmt.Sprintf("FileSystem(AppName=%q, Repo=%q)", fs.appName, fs.repo.URI)
}

func (fs *fileSystem) Create(name string) (File, error) {
	storageName := fs.storageName(name)
	ioErr, grpcErr := fs.client.Storage.Create(fs.ctx, storageName)
	if grpcErr != nil {
		return nil, grpcErr
	}
	if ioErr != nil {
		return nil, storageError(ioErr)
	}
	return &file{
		fs:   fs,
		name: storageName,
	}, nil
}

func (fs *fileSystem) Remove(name string) error {
	ioErr, grpcErr := fs.client.Storage.Remove(fs.ctx, fs.storageName(name))
	if grpcErr != nil {
		return grpcErr
	}
	return storageError(ioErr)
}

func (fs *fileSystem) RemoveAll(name string) error {
	ioErr, grpcErr := fs.client.Storage.RemoveAll(fs.ctx, fs.storageName(name))
	if grpcErr != nil {
		return grpcErr
	}
	return storageError(ioErr)
}

func (fs *fileSystem) Open(name string) (File, error) {
	_, err := fs.Lstat(name)
	if err != nil {
		return nil, err
	}
	return &file{
		fs:   fs,
		name: fs.storageName(name),
	}, nil
}

func (fs *fileSystem) Lstat(path string) (os.FileInfo, error) {
	resp, grpcErr := fs.client.Storage.Stat(fs.ctx, fs.storageName(path))
	if grpcErr != nil {
		return nil, grpcErr
	}
	if resp.Error != nil {
		return nil, storageError(resp.Error)
	}
	return fileInfo{resp.Info}, nil
}

func (fs *fileSystem) ReadDir(path string) ([]os.FileInfo, error) {
	resp, grpcErr := fs.client.Storage.ReadDir(fs.ctx, fs.storageName(path))
	if grpcErr != nil {
		return nil, grpcErr
	}
	if resp.Error != nil {
		return nil, storageError(resp.Error)
	}
	var infos []os.FileInfo
	for _, fi := range resp.Info {
		infos = append(infos, fileInfo{fi})
	}
	return infos, nil
}
