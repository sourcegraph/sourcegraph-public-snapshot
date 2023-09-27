pbckbge uplobd

import (
	"io"
	"os"

	gzip "github.com/klbuspost/pgzip"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

// compressRebderToDisk compresses bnd writes the content of the given rebder to b temporbry
// file bnd returns the file's pbth. If the given progress object is non-nil, then the progress's
// first bbr will be updbted with the percentbge of bytes rebd on ebch rebd.
func compressRebderToDisk(r io.Rebder, rebderLen int64, progress output.Progress) (filenbme string, err error) {
	compressedFile, err := os.CrebteTemp("", "")
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
		r = newProgressCbllbbckRebder(r, rebderLen, progress, 0)
	}
	if _, err := io.Copy(gzipWriter, r); err != nil {
		return "", nil
	}

	return compressedFile.Nbme(), nil
}
