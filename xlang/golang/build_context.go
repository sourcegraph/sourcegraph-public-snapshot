package golang

import (
	"context"
	"go/build"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/ctxvfs"
)

func (h *LangHandler) defaultBuildContext() *build.Context {
	bctx := &build.Default
	if override := h.init.BuildContext; override != nil {
		bctx = &build.Context{
			GOOS:        override.GOOS,
			GOARCH:      override.GOARCH,
			GOPATH:      override.GOPATH,
			GOROOT:      override.GOROOT,
			CgoEnabled:  override.CgoEnabled,
			UseAllFiles: override.UseAllFiles,
			Compiler:    override.Compiler,
			BuildTags:   override.BuildTags,
		}
	}
	return bctx
}

func (h *handlerShared) overlayBuildContext(ctx context.Context, orig *build.Context, useOSFileSystem bool) *build.Context {
	mfs := ctxvfs.Sync(&h.mu, h.fs)

	var fs ctxvfs.FileSystem
	if useOSFileSystem {
		ns := ctxvfs.NameSpace{}
		// The overlay FS takes precedence, but we fall back to the OS
		// file system.
		ns.Bind("/", mfs, "/", ctxvfs.BindReplace)
		ns.Bind("/", ctxvfs.OS("/"), "/", ctxvfs.BindAfter)
		fs = ns
	} else {
		fs = mfs
	}

	return fsBuildContext(ctx, orig, fs)
}

func fsBuildContext(ctx context.Context, orig *build.Context, fs ctxvfs.FileSystem) *build.Context {
	copy := *orig // make a copy
	ctxt := &copy

	ctxt.OpenFile = func(path string) (io.ReadCloser, error) {
		return fs.Open(ctx, path)
	}
	ctxt.IsDir = func(path string) bool {
		fi, err := fs.Stat(ctx, path)
		return err == nil && fi.Mode().IsDir()
	}
	ctxt.HasSubdir = func(root, dir string) (rel string, ok bool) {
		if !ctxt.IsDir(dir) {
			return "", false
		}
		rel, err := filepath.Rel(root, dir)
		if err != nil {
			return "", false
		}
		return rel, true
	}
	ctxt.ReadDir = func(path string) ([]os.FileInfo, error) {
		return fs.ReadDir(ctx, path)
	}
	return ctxt
}

// containingPackage returns the package that contains the given
// file. It is like buildutil.ContainingPackage, except that it
// returns the whole package (i.e., it doesn't use build.FindOnly),
// and it does not perform FS calls that are unnecessary for us (such
// as searching the GOROOT; this is only called on the main
// workspace's code, not its deps).
func containingPackage(bctx *build.Context, filename string) (*build.Package, error) {
	if !strings.HasPrefix(bctx.GOPATH, "/") || strings.Contains(bctx.GOPATH, ":") {
		panic("build context GOPATH must contain exactly 1 entry: " + bctx.GOPATH)
	}

	pkgDir := path.Dir(filename)
	srcDir := path.Join(bctx.GOPATH, "src") + "/"
	importPath := strings.TrimPrefix(pkgDir, srcDir)
	return bctx.Import(importPath, pkgDir, 0)
}
