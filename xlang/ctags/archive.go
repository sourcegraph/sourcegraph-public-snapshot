package ctags

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/neelance/parallel"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/ctxvfs"
)

func copyRepoArchive(ctx context.Context, destination string) error {
	par := parallel.NewRun(10)
	info := ctxInfo(ctx)

	span, ctx := opentracing.StartSpanFromContext(ctx, "copy files from VFS to disk")
	w := ctxvfs.Walk(ctx, "/", info.fs)
	for w.Step() {
		if err := w.Err(); err != nil {
			return err
		}
		fi := w.Stat()
		if boringDir(fi) {
			w.SkipDir()
			continue
		}
		if !fi.Mode().IsRegular() || !isSupportedFile(info.mode, fi.Name()) {
			continue
		}
		par.Acquire()
		go copyFile(ctx, destination, w.Path(), par)
	}
	err := par.Wait()
	span.Finish()
	return err
}

func copyFile(ctx context.Context, destination, path string, par *parallel.Run) {
	var err error
	defer func() {
		if err != nil {
			par.Error(err)
		}
		par.Release()
	}()

	outfile := filepath.Join(destination, path)
	err = os.MkdirAll(filepath.Dir(outfile), os.ModePerm)
	if err != nil {
		return
	}

	in, err := ctxInfo(ctx).fs.Open(ctx, path)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(outfile)
	if err != nil {
		return
	}
	defer out.Close()

	_, err = io.Copy(out, in)
}

func boringDir(fi os.FileInfo) bool {
	if !fi.Mode().IsDir() {
		return false
	}
	switch fi.Name() {
	case ".git", "node_modules", "vendor", "dist", ".srclib-cache":
		return true
	default:
		return false
	}
}
