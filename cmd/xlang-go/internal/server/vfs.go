package server

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/xlang/uri"
	"github.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

var UseRemoteFS = false

// remoteFS returns a virtual file system interface for accessing the files in
// the specified repo at the given commit.
//
// If repo is set and gitserver client is setup, we directly access gitserver
// as a performance and resource optimization. Otherwise we use the FS exposed
// by conn.
//
// SECURITY NOTE: This DOES NOT check that the user or context has permissions
// to read the repo. We assume permission checks happen before a request
// reaches a build server.
func remoteFS(ctx context.Context, conn *jsonrpc2.Conn, workspaceURI lsp.DocumentURI) (ctxvfs.FileSystem, error) {
	if workspaceURI == "" || UseRemoteFS {
		return vfsutil.RemoteFS(conn), nil
	}
	u, err := uri.Parse(string(workspaceURI))
	if err != nil {
		return nil, errors.Wrap(err, "could not parse workspace URI for remotefs")
	}
	if u.Rev() == "" {
		return nil, errors.Errorf("rev is required in uri: %s", workspaceURI)
	}
	archiveFS := vfsutil.NewGitServer(u.Repo(), api.CommitID(u.Rev()))
	archiveFS.EvictOnClose = true
	return archiveFS, nil
}
