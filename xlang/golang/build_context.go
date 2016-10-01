package golang

import (
	"context"
	"go/build"
	"io"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil/lockfs"
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

func (h *handlerShared) overlayBuildContext(orig *build.Context, useOSFileSystem bool) *build.Context {
	mfs := lockfs.New(&h.mu, h.fs)

	var fs vfs.FileSystem
	if useOSFileSystem {
		ns := vfs.NameSpace{}
		// The overlay FS takes precedence, but we fall back to the OS
		// file system.
		ns.Bind("/", mfs, "/", vfs.BindReplace)
		ns.Bind("/", vfs.OS("/"), "/", vfs.BindAfter)
		fs = ns
	} else {
		fs = mfs
	}

	return fsBuildContext(orig, fs)
}

func fsBuildContext(orig *build.Context, fs vfs.FileSystem) *build.Context {
	copy := *orig // make a copy
	ctxt := &copy

	ctxt.OpenFile = func(path string) (io.ReadCloser, error) {
		return fs.Open(path)
	}
	ctxt.IsDir = func(path string) bool {
		fi, err := fs.Stat(path)
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
	ctxt.ReadDir = fs.ReadDir
	return ctxt
}

// containingPackage returns the package that contains the given
// file. It is like buildutil.ContainingPackage, except that it
// returns the whole package (i.e., it doesn't use build.FindOnly),
// and it does not perform FS calls that are unnecessary for us (such
// as searching the GOROOT; this is only called on the main
// workspace's code, not its deps).
func containingPackage(ctx context.Context, bctx *build.Context, filename string) (*build.Package, error) {
	if !strings.HasPrefix(bctx.GOPATH, "/") || strings.Contains(bctx.GOPATH, ":") {
		panic("build context GOPATH must contain exactly 1 entry: " + bctx.GOPATH)
	}

	pkgDir := path.Dir(filename)
	srcDir := path.Join(bctx.GOPATH, "src") + "/"
	importPath := strings.TrimPrefix(pkgDir, srcDir)
	return bctx.Import(importPath, pkgDir, 0)
}
