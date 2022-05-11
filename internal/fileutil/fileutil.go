package fileutil

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

// UpdateFileIfDifferent will atomically update the file if the contents are
// different. If it does an update ok is true.
func UpdateFileIfDifferent(path string, content []byte) (bool, error) {
	current, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		// If the file doesn't exist we write a new file.
		return false, err
	}

	if bytes.Equal(current, content) {
		return false, nil
	}

	// We write to a tempfile first to do the atomic update (via rename)
	f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path))
	if err != nil {
		return false, err
	}
	// We always remove the tempfile. In the happy case it won't exist.
	defer os.Remove(f.Name())

	if n, err := f.Write(content); err != nil {
		f.Close()
		return false, err
	} else if n != len(content) {
		f.Close()
		return false, io.ErrShortWrite
	}

	// fsync to ensure the disk contents are written. This is important, since
	// we are not guaranteed that os.Rename is recorded to disk after f's
	// contents.
	if err := f.Sync(); err != nil {
		f.Close()
		return false, err
	}
	if err := f.Close(); err != nil {
		return false, err
	}
	return true, RenameAndSync(f.Name(), path)
}

// RenameAndSync will do an os.Rename followed by fsync to ensure the rename
// is recorded
func RenameAndSync(oldpath, newpath string) error {
	err := os.Rename(oldpath, newpath)
	if err != nil {
		return err
	}

	oldparent, newparent := filepath.Dir(oldpath), filepath.Dir(newpath)
	err = fsync(newparent)
	if oldparent != newparent {
		if err1 := fsync(oldparent); err == nil {
			err = err1
		}
	}
	return err
}

func fsync(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
