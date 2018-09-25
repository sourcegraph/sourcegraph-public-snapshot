package proxy

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/xlang/uri"
)

// relWorkspaceURI maps absolute URIs like
// "git://github.com/facebook/react.git?master#dir/file.txt" to
// workspace-relative file URIs like "file:///dir/file.txt". The
// result is a path within the workspace's virtual file system that
// will contain the original path's contents.
//
// If uriStr isn't underneath root, the parsed original uriStr is
// returned. This occurs when a cross-workspace resource is referenced
// (e.g., a client performs a cross-workspace go-to-definition and
// then notifies the server that the client opened the destination
// resource).
func relWorkspaceURI(root uri.URI, uriStr string) (*uri.URI, error) {
	u, err := uri.Parse(uriStr)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(uriStr, root.String()) {
		// The uriStr resource lives in a different workspace.
		return u, nil
	}
	if p := path.Clean(u.FilePath()); strings.HasPrefix(p, "/") || strings.HasPrefix(p, "..") {
		return nil, fmt.Errorf("invalid file path in URI %q in LSP proxy client request (must not begin with '/', '..', or contain '.' or '..' components)", uriStr)
	} else if u.FilePath() != "" && p != u.FilePath() {
		return nil, fmt.Errorf("invalid file path in URI %q (raw file path %q != cleaned file path %q)", uriStr, u.FilePath(), p)
	}

	// Support when root is rooted at a subdir.
	if rootPath := root.FilePath(); rootPath != "" {
		rootPath = strings.TrimSuffix(rootPath, string(os.PathSeparator))
		if !strings.HasPrefix(u.FilePath(), rootPath+string(os.PathSeparator)) {
			return u, nil
		}
		u = u.WithFilePath(strings.TrimPrefix(u.FilePath(), rootPath+string(os.PathSeparator)))
	}

	return &uri.URI{URL: url.URL{Scheme: "file", Path: "/" + u.FilePath()}}, nil
}

// absWorkspaceURI is the inverse of relWorkspaceURI. It maps
// workspace-relative URIs like "file:///dir/file.txt" to their
// absolute URIs like
// "git://github.com/facebook/react.git?master#dir/file.txt".
func absWorkspaceURI(root uri.URI, uriStr string) (*uri.URI, error) {
	uri, err := uri.Parse(uriStr)
	if err != nil {
		return nil, err
	}
	if uri.Scheme == "file" {
		return root.WithFilePath(root.ResolveFilePath(uri.Path)), nil
	}
	return uri, nil
	// Another possibility is a "git://" URI that the build/lang
	// server knew enough to produce on its own (e.g., to refer to
	// git://github.com/golang/go for a Go stdlib definition). No need
	// to rewrite those.
}
