package golang

import (
	"go/build"
	"io"
	"path/filepath"

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
