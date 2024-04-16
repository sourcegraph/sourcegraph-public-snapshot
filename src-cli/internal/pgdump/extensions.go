package pgdump

import (
	"bufio"
	"bytes"
	"io"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// PartialCopyWithoutExtensions will perform a partial copy of a SQL database dump from
// src to dst while commenting out EXTENSIONs-related statements. When it determines there
// are no more EXTENSIONs-related statements, it will return, resetting src to the position
// of the last contents written to dst.
//
// This is needed for import to Google Cloud Storage, which does not like many EXTENSION
// statements. For more details, see https://cloud.google.com/sql/docs/postgres/import-export/import-export-dmp
//
// Filtering requires reading entire lines into memory - this can be a very expensive
// operation, so when filtering is complete the more efficient io.Copy should be used
// to perform the remainder of the copy from src to dst.
func PartialCopyWithoutExtensions(dst io.Writer, src io.ReadSeeker, progressFn func(int64)) (int64, error) {
	var (
		reader = bufio.NewReader(src)
		// position we have consumed up to, track separately because bufio.Reader may have
		// read ahead on src. This allows us to reset src later.
		consumed int64
		// number of bytes we have actually written to dst - it should always be returned.
		written int64
		// set to true when we have done all our filtering
		noMoreExtensions bool
	)

	for !noMoreExtensions {
		// Read up to a line, keeping track of our position in src
		line, err := reader.ReadBytes('\n')
		consumed += int64(len(line))
		if err != nil {
			return written, err
		}

		// Once we start seeing table creations, we are definitely done with extensions,
		// so we can hand off the rest to the superior io.Copy implementation.
		if bytes.HasPrefix(line, []byte("CREATE TABLE")) {
			// we are done with extensions
			noMoreExtensions = true
		} else if bytes.HasPrefix(line, []byte("COMMENT ON EXTENSION")) {
			// comment out this line
			line = append([]byte("-- "), line...)
		}

		// Write this line and update our progress before returning on error
		lineWritten, err := dst.Write(line)
		written += int64(lineWritten)
		progressFn(written)
		if err != nil {
			return written, err
		}
	}

	// No more extensions - reset src to the last actual consumed position
	_, err := src.Seek(consumed, io.SeekStart)
	if err != nil {
		return written, errors.Wrap(err, "reset src position")
	}
	return written, nil
}
