package xlang

import (
	"context"
	"net/url"
	"strings"

	"github.com/sourcegraph/ctxvfs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
)

// NewRemoteRepoVFS returns a virtual file system interface for
// accessing the files in the specified repo at the given commit.
//
// SECURITY NOTE: NewRemoteRepoVFS DOES NOT check that the user or
// context has permissions to read the repo. The permission check must
// be performed by the caller to the LSP client proxy.
//
// It is a var so that it can be mocked in tests.
var NewRemoteRepoVFS = func(ctx context.Context, cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
	repo := cloneURL.Host + strings.TrimSuffix(cloneURL.Path, ".git")
	return vcs.FastVFS(ctx, gitcmd.Open(repo), rev)
}
