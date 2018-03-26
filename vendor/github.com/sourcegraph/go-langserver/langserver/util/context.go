package util

import (
	"go/build"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/sourcegraph/ctxvfs"

	"golang.org/x/net/context"
)

func PrepareContext(bctx *build.Context, ctx context.Context, fs ctxvfs.FileSystem) {
	// HACK: in the all Context's methods below we are trying to convert path to virtual one (/foo/bar/..)
	// because some code may pass OS-specific arguments.
	// See golang.org/x/tools/go/buildutil/allpackages.go which uses `filepath` for example

	bctx.OpenFile = func(path string) (io.ReadCloser, error) {
		return fs.Open(ctx, virtualPath(path))
	}
	bctx.IsDir = func(path string) bool {
		fi, err := fs.Stat(ctx, virtualPath(path))
		return err == nil && fi.Mode().IsDir()
	}
	bctx.HasSubdir = func(root, dir string) (rel string, ok bool) {
		if !bctx.IsDir(dir) {
			return "", false
		}
		if !PathHasPrefix(dir, root) {
			return "", false
		}
		return PathTrimPrefix(dir, root), true
	}
	bctx.ReadDir = func(path string) ([]os.FileInfo, error) {
		return fs.ReadDir(ctx, virtualPath(path))
	}
	bctx.IsAbsPath = func(path string) bool {
		return IsAbs(path)
	}
	bctx.JoinPath = func(elem ...string) string {
		// convert all backslashes to slashes to avoid
		// weird paths like C:\mygopath\/src/github.com/...
		for i, el := range elem {
			elem[i] = filepath.ToSlash(el)
		}
		return path.Join(elem...)
	}
}
