package server

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
	"github.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

// RemoteFS fetches a zip archive from gitserver and returns a virtual file
// system interface for accessing the files in the specified repo at the given
// commit.
//
// SECURITY NOTE: This DOES NOT check that the user or context has permissions
// to read the repo. We assume permission checks happen before a request reaches
// a build server.
var RemoteFS = func(ctx context.Context, conn *jsonrpc2.Conn, workspaceURI lsp.DocumentURI) (ctxvfs.FileSystem, error) {
	u, err := gituri.Parse(string(workspaceURI))
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
