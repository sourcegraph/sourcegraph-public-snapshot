package upload

import (
	"io"
	"os"

	gzip "github.com/klauspost/pgzip"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// compressReaderToDisk compresses and writes the content of the given reader to a temporary
// file and returns the file's path. If the given progress object is non-nil, then the progress's
// first bar will be updated with the percentage of bytes read on each read.
func compressReaderToDisk(r io.Reader, readerLen int64, progress output.Progress) (filename string, err error) {
	compressedFile, err := os.CreateTemp("", "")
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := compressedFile.Close(); err != nil {
			err = errors.Append(err, closeErr)
		}
	}()

	gzipWriter := gzip.NewWriter(compressedFile)
	defer func() {
		if closeErr := gzipWriter.Close(); err != nil {
			err = errors.Append(err, closeErr)
		}
	}()

	if progress != nil {
		r = newProgressCallbackReader(r, readerLen, progress, 0)
	}
	if _, err := io.Copy(gzipWriter, r); err != nil {
		return "", nil
	}

	return compressedFile.Name(), nil
}
