package vfsutil

import (
	"context"
	"net/url"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/ctxvfs"
)

// NewRemoteRepoVFS returns a virtual file system interface for
// accessing the files in the specified repo at the given commit.
//
// It is a var so that it can be mocked in tests.
var NewRemoteRepoVFS = func(ctx context.Context, cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
	// Fast-path for GitHub repos, which we can fetch on-demand from
	// GitHub's repo .zip archive download endpoint.
	if cloneURL.Host == "github.com" {
		fullName := cloneURL.Host + strings.TrimSuffix(cloneURL.Path, ".git") // of the form "github.com/foo/bar"
		return NewGitHubRepoVFS(fullName, rev, "", true)
	}

	// Fall back to a full git clone for non-github.com repos.
	return &GitRepoVFS{
		CloneURL: cloneURL.String(),
		Rev:      rev,
	}, nil
}
