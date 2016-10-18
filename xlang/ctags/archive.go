package ctags

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/ctxvfs"
	"golang.org/x/net/context"
)

func copyRepoArchive(ctx context.Context, fs ctxvfs.FileSystem, destination string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "read files from network")
	files, err := ctxvfs.ReadAllFiles(ctx, fs, "", nil)
	span.Finish()
	if err != nil {
		return err
	}

	span, ctx = opentracing.StartSpanFromContext(ctx, "write files to disk")
	defer span.Finish()
	for relpath, contents := range files {
		filename := filepath.Clean(filepath.Join(destination, relpath))
		dirname := strings.TrimSuffix(filename, filepath.Base(filename))

		if err := os.MkdirAll(dirname, os.ModePerm); err != nil {
			return err
		}

		err := ioutil.WriteFile(filename, contents, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
