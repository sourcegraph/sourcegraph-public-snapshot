package indexer

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
)

func extractTarfile(root string, r io.Reader) error {
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

func extractFile(root string, tr *tar.Reader, header *tar.Header) error {
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
	defer f.Close()

	if _, err := io.Copy(f, tr); err != nil {
		return err
	}

	return nil
}
