package codeintel

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/hashicorp/go-multierror"
)

// splitFile writes the contents of the given file into a series of temporary
// files, each of which are gzipped and are no larger than the given max payload
// size. The file names are returned in the order in which they were written. The
// cleanup function removes all temporary files, and wraps the error argument
// with any additional errors that happen during cleanup.
func splitFile(file string, maxPayloadSize int) (files []string, _ func(error) error, err error) {
	cleanup := func(err error) error {
		for _, file := range files {
			if removeErr := os.Remove(file); removeErr != nil {
				err = multierror.Append(err, removeErr)
			}
		}

		return err
	}
	defer func() {
		if err != nil {
			err = cleanup(err)
		}
	}()

	sourceFile, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}
	defer sourceFile.Close()

	for {
		partFile, err := ioutil.TempFile("", "")
		if err != nil {
			return nil, nil, err
		}
		defer partFile.Close()

		n, err := io.CopyN(partFile, sourceFile, int64(maxPayloadSize))
		if err != nil && err != io.EOF {
			return nil, nil, err
		}

		if n > 0 {
			files = append(files, partFile.Name())
		} else {
			// File must be closed before it's removed (on Windows), so
			// don't wait for the defer above. Double closing a file is
			// fine, but the second close will return an ErrClosed value,
			// which we aren't checking anyway.
			_ = partFile.Close()

			// Edge case: previous io.CopyN call returned err=nil and
			// n=maxPayloadSize. Nothing written to this file so we
			// can just undo its creation and fall-through to the break.
			if err := os.Remove(partFile.Name()); err != nil {
				return nil, nil, err
			}
		}

		if err == io.EOF {
			break
		}
	}

	return files, cleanup, nil
}
