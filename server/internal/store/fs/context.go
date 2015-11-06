package fs

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/rwvfs"
)

type contextKey int

const (
	reposAbsPathKey contextKey = iota
	buildStoreVFSKey
	usersVFSKey
	repoStatusVFSKey
	appStorageDirKey
)

// WithReposVFS creates a new child context that looks for
// repositories at the root of the given path.
func WithReposVFS(parent context.Context, absPath string) context.Context {
	return context.WithValue(parent, reposAbsPathKey, absPath)
}

// WithBuildStoreVFS creates a new child context that looks for
// the build store at the root of the given VFS.
func WithBuildStoreVFS(parent context.Context, fs rwvfs.FileSystem) context.Context {
	return context.WithValue(parent, buildStoreVFSKey, rwvfs.Walkable(fs))
}

// WithDBVFS creates a new child context that reads and writes general
// data (users, registered API clients, passwords, etc.) in fs.
func WithDBVFS(parent context.Context, fs rwvfs.FileSystem) context.Context {
	return context.WithValue(parent, usersVFSKey, fs)
}

// WithRepoStatusVFS creates a new child context that reads and writes
// repo status data.
func WithRepoStatusVFS(parent context.Context, fs rwvfs.FileSystem) context.Context {
	return context.WithValue(parent, repoStatusVFSKey, rwvfs.Walkable(fs))
}

// WithAppStorageDir creates a new child context that reads and writes app
// storage data at the given directory.
func WithAppStorageDir(parent context.Context, dir string) context.Context {
	return context.WithValue(parent, appStorageDirKey, dir)
}

// reposAbsPath returns the absolute path of the repository storage directory.
func reposAbsPath(ctx context.Context) string {
	return mustString(ctx, reposAbsPathKey)
}

// buildStoreVFS returns a virtual filesystem pointed to where build data is stored.
func buildStoreVFS(ctx context.Context) rwvfs.WalkableFileSystem {
	return mustWalkableVFS(ctx, buildStoreVFSKey)
}

// repoStatusVFS returns the virtual filesystem pointed to where repo
// statuses are stored.
func repoStatusVFS(ctx context.Context) rwvfs.WalkableFileSystem {
	return mustWalkableVFS(ctx, repoStatusVFSKey)
}

// appStorageDir returns the virtual filesystem pointed to where app storage is
// located.
func appStorageDir(ctx context.Context) string {
	return mustString(ctx, appStorageDirKey)
}

// dbVFS returns the VFS in which most other data is stored (users,
// registered API clients, passwords, etc.).
func dbVFS(ctx context.Context) rwvfs.FileSystem {
	return mustVFS(ctx, usersVFSKey)
}

func mustString(ctx context.Context, key contextKey) string {
	str, ok := ctx.Value(key).(string)
	if !ok {
		panic("no repos absolute path set in context")
	}
	return str
}

func mustVFS(ctx context.Context, key contextKey) rwvfs.FileSystem {
	vfs, ok := ctx.Value(key).(rwvfs.FileSystem)
	if !ok {
		panic("no FileSystem set in context")
	}
	return vfs
}

func mustWalkableVFS(ctx context.Context, key contextKey) rwvfs.WalkableFileSystem {
	vfs, ok := ctx.Value(key).(rwvfs.WalkableFileSystem)
	if !ok {
		panic("no WalkableFileSystem set in context")
	}
	return vfs
}
