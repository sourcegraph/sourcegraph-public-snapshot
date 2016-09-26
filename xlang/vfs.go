package xlang

import (
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil"

	"golang.org/x/tools/godoc/vfs"
)

func CreateGitVFS(urlStr string) (vfs.FileSystem, error) {
	u, err := uri.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	return VFSCreatorsByScheme["git"](u)
}

var VFSCreatorsByScheme = map[string]func(root *uri.URI) (vfs.FileSystem, error){
	"git": func(root *uri.URI) (vfs.FileSystem, error) {
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
