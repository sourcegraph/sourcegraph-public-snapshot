package tar

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
)

// Extract reads tar archive data from r and extracts it into files under the given root.
func Extract(root string, r io.Reader) error {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		if err := extractHeader(root, tr, header); err != nil {
			return err
		}
	}

	return nil
}

func extractHeader(root string, tr *tar.Reader, header *tar.Header) error {
	switch header.Typeflag {
	case tar.TypeDir:
		return extractDir(root, tr, header)
	case tar.TypeReg:
		return extractFile(root, tr, header)
	}

	return nil
}

func extractDir(root string, tr *tar.Reader, header *tar.Header) error {
	target := filepath.Join(root, header.Name)

	if _, err := os.Stat(target); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := os.MkdirAll(target, 0755); err != nil {
			return err
		}
	}

	return nil
}

func extractFile(root string, tr *tar.Reader, header *tar.Header) (err error) {
	target := filepath.Join(root, header.Name)

	// It's possible for a file to exist in a directory for which there is
	// no header. This happens when creating a tar with `tar -cvf dir/*`.
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	_, err = io.Copy(f, tr)
	return err
}
