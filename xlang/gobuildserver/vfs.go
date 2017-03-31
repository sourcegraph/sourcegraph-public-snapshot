package gobuildserver

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/jsonrpc2"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

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
func remoteFS(ctx context.Context, conn *jsonrpc2.Conn, workspaceURI string) (ctxvfs.FileSystem, error) {
	if workspaceURI == "" || !gitserver.DefaultClient.HasServers() {
		return vfsutil.RemoteFS(conn), nil
	}
	u, err := uri.Parse(workspaceURI)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse workspace URI for remotefs")
	}
	if u.Rev() == "" {
		return nil, errors.Errorf("rev is required in uri: %s", workspaceURI)
	}
	return vcs.ArchiveFileSystem(gitcmd.Open(&sourcegraph.Repo{URI: u.Repo()}), u.Rev()), nil
}
