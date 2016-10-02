package xlang

import (
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/ctxvfs"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

func CreateGitVFS(urlStr string) (ctxvfs.FileSystem, error) {
	u, err := uri.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	return VFSCreatorsByScheme["git"](u)
}

var VFSCreatorsByScheme = map[string]func(root *uri.URI) (ctxvfs.FileSystem, error){
	"git": func(root *uri.URI) (ctxvfs.FileSystem, error) {
		if root.Host == "github.com" {
			return vfsutil.NewGitHubRepoVFS(root.Host+root.Path, root.Rev(), root.FilePath(), true)
		}

		// Fall back to git clone.
		return &vfsutil.GitRepoVFS{
			CloneURL: root.CloneURL().String(),
			Rev:      root.Rev(),
			Subtree:  root.FilePath(),
		}, nil
	},
}
