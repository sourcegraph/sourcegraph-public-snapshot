package server

import (
	"compress/gzip"
	"io"
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
)

// stitchMultipart combines multiple uploads into a single compressed file. Each part on disk
// with the name `{bundleId}.{partIndex}.lsif.gz` will be concatenated into a single bundle
// file with the name `bundleId.lsif.gz`. The content of each part is decompressed and written
// to a new compressed file sequentially. On success, the part files are removed.
func stitchMultipart(bundleDir string, id int64) error {
	targetFile, err := os.OpenFile(paths.UploadFilename(bundleDir, id), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := targetFile.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	writer := gzip.NewWriter(targetFile)
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	index := 0
	for {
		exists, reader, err := openPart(bundleDir, id, int64(index))
		if err != nil {
			return err
		}
		if !exists {
			break
		}
		defer reader.Close()

		if _, err := io.Copy(writer, reader); err != nil {
			return err
		}

		index++
	}

	for i := index - 1; i >= 0; i-- {
		if err := os.Remove(paths.UploadPartFilename(bundleDir, id, int64(i))); err != nil {
			log15.Error("Failed to remove bundle part", "bundleID", id, "index", index, "err", err)
		}
	}

	return nil
}

// openPart opens a gzip reader for a upload part file as well as a boolean flag
// indicating if the file exists.
func openPart(bundleDir string, id, index int64) (bool, io.ReadCloser, error) {
	f, err := os.Open(paths.UploadPartFilename(bundleDir, id, int64(index)))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil, nil
		}

		return false, nil, err
	}

	reader, err := gzip.NewReader(f)
	if err != nil {
		return false, nil, err
	}

	return true, &partReader{reader, f}, nil
}

// partReader bundles a gzip reader with its underlying reader. This overrides the
// Close method on the gzip reader so that it also closes the underlying reader.
type partReader struct {
	*gzip.Reader
	rc io.ReadCloser
}

func (r *partReader) Close() error {
	for _, err := range []error{r.Reader.Close(), r.rc.Close()} {
		if err != nil {
			return err
		}
	}

	return nil
}
