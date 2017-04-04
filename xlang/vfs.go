package xlang

import (
	"context"
	"net/url"
	"strings"
	"sync"

	"github.com/sourcegraph/ctxvfs"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

// NewRemoteRepoVFS returns a virtual file system interface for
// accessing the files in the specified repo at the given commit.
//
// SECURITY NOTE: NewRemoteRepoVFS DOES NOT check that the user or
// context has permissions to read the repo. The permission check must
// be performed by the caller to the LSP client proxy.
//
// It is a var so that it can be mocked in tests.
var NewRemoteRepoVFS = func(ctx context.Context, cloneURL *url.URL, rev string) (FileSystem, error) {
	repo := cloneURL.Host + strings.TrimSuffix(cloneURL.Path, ".git")
	key := sharedFSKey{repo: repo, rev: rev}

	// Share an open archive amongst clients. It is common to open a repo
	// more than once, since we open it up per mode.
	openSharedFSMu.Lock()
	defer openSharedFSMu.Unlock()
	fs, ok := openSharedFS[key]
	if !ok {
		fs = &sharedFS{
			// ArchiveFileSystem and gitcmd.Open do not block, so
			// we can create them while holding the lock.
			FileSystem: vfsutil.ArchiveFileSystem(gitcmd.Open(&sourcegraph.Repo{URI: repo}), rev),
			key:        key,
		}
		openSharedFS[key] = fs
	}
	fs.numReaders++
	return fs, nil
}

type FileSystem interface {
	ctxvfs.FileSystem
	ListAllFiles(ctx context.Context) ([]string, error)
}

var (
	openSharedFSMu sync.Mutex
	openSharedFS   map[sharedFSKey]*sharedFS
)

func init() {
	openSharedFS = make(map[sharedFSKey]*sharedFS)
}

type sharedFSKey struct {
	repo, rev string
}

// sharedFS tracks multiple readers to a FileSystem. This allows us to not
// fetch a repository multiple times (once for each active mode).
type sharedFS struct {
	FileSystem
	key        sharedFSKey
	numReaders int
}

func (fs *sharedFS) Close() error {
	openSharedFSMu.Lock()
	fs.numReaders--
	if fs.numReaders == 0 {
		delete(openSharedFS, fs.key)
	}
	openSharedFSMu.Unlock()
	return nil
}
