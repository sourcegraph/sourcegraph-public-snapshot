package langserver

import (
	"fmt"
	"go/build"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"

	"golang.org/x/tools/go/buildutil"

	"github.com/sourcegraph/go-langserver/langserver/util"
)

// BuildContext creates a build.Context which uses the overlay FS and the InitializeParams.BuildContext overrides.
func (h *LangHandler) BuildContext(ctx context.Context) *build.Context {
	var bctx *build.Context
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
	} else {
		// make a copy since we will mutate it
		copy := build.Default
		bctx = &copy
	}

	h.Mu.Lock()
	fs := h.FS
	h.Mu.Unlock()

	util.PrepareContext(bctx, ctx, fs)
	return bctx
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
	gopaths := buildutil.SplitPathList(bctx, bctx.GOPATH) // list will be empty with no GOPATH
	for _, gopath := range gopaths {
		if !buildutil.IsAbsPath(bctx, gopath) {
			return nil, fmt.Errorf("build context GOPATH must be an absolute path (GOPATH=%q)", gopath)
		}
	}

	pkgDir := filename
	if !bctx.IsDir(filename) {
		pkgDir = path.Dir(filename)
	}
	var srcDir string
	if util.PathHasPrefix(filename, bctx.GOROOT) {
		srcDir = bctx.GOROOT // if workspace is Go stdlib
	} else {
		srcDir = "" // with no GOPATH, only stdlib will work
		for _, gopath := range gopaths {
			if util.PathHasPrefix(pkgDir, gopath) {
				srcDir = gopath
				break
			}
		}
	}
	srcDir = path.Join(filepath.ToSlash(srcDir), "src")
	importPath := util.PathTrimPrefix(pkgDir, srcDir)
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
