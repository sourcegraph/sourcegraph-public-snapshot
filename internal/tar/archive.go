package tar

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
)

// Archive walks the files rooted at the given path and streams them to a tar archive
// contained in the resulting reader. Any errors that occur while reading files on disk
// are exposed through Read calls on the the resulting reader.
func Archive(root string) io.Reader {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		tw := tar.NewWriter(pw)
		defer tw.Close()

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			header.Name = strings.TrimPrefix(strings.TrimPrefix(path, root), string(filepath.Separator))

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			return archiveFile(tw, path)
		})
		if err != nil {
			_ = pw.CloseWithError(err)
		}
	}()

	return pr
}

func archiveFile(w io.Writer, filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	_, err = io.Copy(w, f)
	return err
}
