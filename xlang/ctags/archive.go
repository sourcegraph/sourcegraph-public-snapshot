package ctags

import (
	"io"
	"os"
	"path/filepath"

	"github.com/neelance/parallel"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/ctxvfs"
	"golang.org/x/net/context"
)

func copyRepoArchive(ctx context.Context, fs ctxvfs.FileSystem, destination string) error {
	filter := func(f os.FileInfo) bool {
		return isSupportedFile(f.Name())
	}

	par := parallel.NewRun(10)

	span, ctx := opentracing.StartSpanFromContext(ctx, "copy files from VFS to disk")
	w := ctxvfs.Walk(ctx, "/", fs)
	for w.Step() {
		if err := w.Err(); err != nil {
			return err
		}
		fiPath := w.Path()
		fi := w.Stat()
		switch {
		case fi.Name() == ".git" && fi.Mode().IsDir():
			w.SkipDir()
		case fi.Mode().IsRegular():
			if filter == nil || filter(fi) {
				par.Acquire()
				go func() {
					defer par.Release()

					outfile := filepath.Join(destination, fiPath)
					if err := os.MkdirAll(filepath.Dir(outfile), os.ModePerm); err != nil {
						par.Error(err)
						return
					}

					in, err := fs.Open(ctx, fiPath)
					if err != nil {
						par.Error(err)
						return
					}

					out, err := os.Create(outfile)
					if err != nil {
						par.Error(err)
						return
					}
					defer out.Close()

					if _, err := io.Copy(out, in); err != nil {
						par.Error(err)
						return
					}
				}()
			}
		}
	}
	err := par.Wait()
	span.Finish()
	if err != nil {
		return err
	}

	return nil
}
