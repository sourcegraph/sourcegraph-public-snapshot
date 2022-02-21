package api

import (
	"compress/gzip"
	"io"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// gzipReader decorates a source reader by gzip compressing its contents.
func gzipReader(source io.Reader) io.Reader {
	r, w := io.Pipe()
	go func() {
		// propagate gzip write errors into new reader
		w.CloseWithError(gzipPipe(source, w))
	}()
	return r
}

// gzipPipe reads uncompressed data from r and writes compressed data to w.
func gzipPipe(r io.Reader, w io.Writer) (err error) {
	gzipWriter := gzip.NewWriter(w)
	defer func() {
		if closeErr := gzipWriter.Close(); closeErr != nil {
			err = errors.Append(err, err)
		}
	}()

	_, err = io.Copy(gzipWriter, r)
	return err
}
