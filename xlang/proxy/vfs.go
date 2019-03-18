package proxy

import (
	"context"
	"io"
	"net/url"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf/reposource"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

// NewRemoteRepoVFS returns a virtual file system interface for
// accessing the files in the specified repo at the given commit.
//
// SECURITY NOTE: NewRemoteRepoVFS DOES NOT check that the user or
// context has permissions to read the repo. The permission check must
// be performed by the caller to the LSP client proxy.
//
// It is a var so that it can be mocked in tests.
var NewRemoteRepoVFS = func(ctx context.Context, cloneURL *url.URL, commitID api.CommitID) (FileSystem, error) {
	repo, err := reposource.CloneURLToRepoURI(cloneURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "can't determine repo name for URL")
	}
	if repo == "" {
		// best effort fallback
		repo = api.RepoURI(cloneURL.Host + strings.TrimSuffix(cloneURL.Path, ".git"))
	}
	// We can get to this point without checking if (repo, commit) actually
	// exists. Its better to fail sooner, otherwise the error can cause a
	// later process to fail (since ArchiveFS fetches lazily). So we check
	// existence first.
	_, err = git.ResolveRevision(ctx, gitserver.Repo{Name: repo}, nil, string(commitID), &git.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return nil, err
	}

	key := sharedFSKey{repo: repo, commitID: commitID}

	// Share an open archive amongst clients. It is common to open a repo
	// more than once, since we open it up per mode.
	openSharedFSMu.Lock()
	defer openSharedFSMu.Unlock()
	fs, ok := openSharedFS[key]
	if !ok {
		// Since we are wrapped by sharedFS, we will only call
		// Close when all readers are done. When that happens
		// we likely won't open the archive again soon, so we
		// can reclaim the disk space it uses.
		archiveFS := vfsutil.NewGitServer(repo, commitID)
		archiveFS.EvictOnClose = true
		fs = &sharedFS{
			FileSystem: archiveFS,
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
	repo     api.RepoURI
	commitID api.CommitID
}

// sharedFS tracks multiple readers to a FileSystem. This allows us to not
// fetch a repository multiple times (once for each active mode).
type sharedFS struct {
	FileSystem
	key        sharedFSKey
	numReaders int
}

func (fs *sharedFS) Close() error {
	close := false
	openSharedFSMu.Lock()
	fs.numReaders--
	if fs.numReaders == 0 {
		close = true
		delete(openSharedFS, fs.key)
	}
	openSharedFSMu.Unlock()
	if close {
		if closer, ok := fs.FileSystem.(io.Closer); ok {
			return closer.Close()
		}
	}
	return nil
}
