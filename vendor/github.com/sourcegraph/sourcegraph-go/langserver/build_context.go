package langserver

import (
	"context"
	"go/build"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/ctxvfs"
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

func (h *HandlerShared) OverlayBuildContext(ctx context.Context, orig *build.Context, useOSFileSystem bool) *build.Context {
	fs := ctxvfs.FileSystem(h.FS)
	if h.AugmentFileSystem != nil {
		fs = h.AugmentFileSystem(fs)
	}
	fs = ctxvfs.Sync(&h.Mu, fs) // protect against race conditions when new binds are mounted
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

// ContainingPackage returns the package that contains the given
// file. It is like buildutil.ContainingPackage, except that:
//
// * it returns the whole package (i.e., it doesn't use build.FindOnly)
// * it does not perform FS calls that are unnecessary for us (such
//   as searching the GOROOT; this is only called on the main
//   workspace's code, not its deps).
// * if the file is in the xtest package (package p_test not package p),
//   it returns build.Package only representing that xtest package
func ContainingPackage(bctx *build.Context, filename string) (*build.Package, error) {
	if !strings.HasPrefix(bctx.GOPATH, "/") || strings.Contains(bctx.GOPATH, ":") {
		panic("build context GOPATH must contain exactly 1 entry: " + bctx.GOPATH)
	}

	pkgDir := path.Dir(filename)
	var srcDir string
	if PathHasPrefix(filename, bctx.GOROOT) {
		srcDir = bctx.GOROOT // if workspace is Go stdlib
	} else {
		srcDir = bctx.GOPATH
	}
	srcDir = path.Join(srcDir, "src") + "/"
	importPath := strings.TrimPrefix(pkgDir, srcDir)
	var xtest bool
	pkg, err := bctx.Import(importPath, pkgDir, 0)
	if pkg != nil {
		base := path.Base(filename)
		for _, f := range pkg.XTestGoFiles {
			if f == base {
				xtest = true
				break
			}
		}
	}

	// If the filename we want refers to a file in an xtest package
	// (package p_test not package p), then munge the package so that
	// it only refers to that xtest package.
	if pkg != nil && xtest && !strings.HasSuffix(pkg.Name, "_test") {
		pkg.Name += "_test"
		pkg.GoFiles = nil
		pkg.CgoFiles = nil
		pkg.TestGoFiles = nil
	}

	return pkg, err
}
