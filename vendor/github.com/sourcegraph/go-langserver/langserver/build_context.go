package langserver

import (
	"context"
	"fmt"
	"go/build"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
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

			// Enable analysis of all go version build tags that
			// our compiler should understand.
			ReleaseTags: build.Default.ReleaseTags,
		}
	}
	return bctx
}

func (h *HandlerShared) OverlayBuildContext(ctx context.Context, orig *build.Context, useOSFileSystem bool) *build.Context {
	// Copy to avoid a race condition if concurrent goroutines
	// bind new mountpoints to the ctxvfs.NameSpace. This lets us lock
	// once here, instead of once per call to any ctxvfs.FileSystem
	// method.
	h.Mu.Lock()
	fs := make(ctxvfs.NameSpace, len(h.FS))
	for path, mounts := range h.FS {
		fs[path] = mounts
	}
	h.Mu.Unlock()

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

// From: https://github.com/golang/tools/blob/b814a3b030588c115189743d7da79bce8b549ce1/go/buildutil/util.go#L84
// dirHasPrefix tests whether the directory dir begins with prefix.
func dirHasPrefix(dir, prefix string) bool {
	if runtime.GOOS != "windows" {
		return strings.HasPrefix(dir, prefix)
	}
	return len(dir) >= len(prefix) && strings.EqualFold(dir[:len(prefix)], prefix)
}

// ContainingPackage returns the package that contains the given
// filename. It is like buildutil.ContainingPackage, except that:
//
// * it returns the whole package (i.e., it doesn't use build.FindOnly)
// * it does not perform FS calls that are unnecessary for us (such
//   as searching the GOROOT; this is only called on the main
//   workspace's code, not its deps).
// * if the file is in the xtest package (package p_test not package p),
//   it returns build.Package only representing that xtest package
func ContainingPackage(bctx *build.Context, filename string) (*build.Package, error) {
	gopaths := filepath.SplitList(bctx.GOPATH) // list will be empty with no GOPATH
	for _, gopath := range gopaths {
		if !filepath.IsAbs(gopath) {
			return nil, fmt.Errorf("build context GOPATH must be an absolute path (GOPATH=%q)", gopath)
		}
	}

	pkgDir := filename
	if !bctx.IsDir(filename) {
		pkgDir = path.Dir(filename)
	}
	var srcDir string
	if PathHasPrefix(filename, bctx.GOROOT) {
		srcDir = bctx.GOROOT // if workspace is Go stdlib
	} else {
		srcDir = "" // with no GOPATH, only stdlib will work
		for _, gopath := range gopaths {
			if dirHasPrefix(pkgDir, gopath) {
				srcDir = gopath
				break
			}
		}
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
