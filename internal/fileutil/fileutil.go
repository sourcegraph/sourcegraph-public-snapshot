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
	// preserve permissions
	// silently ignore failure to avoid breaking changes
	if fileInfo, err := os.Stat(path); err == nil {
		_ = os.Chmod(f.Name(), fileInfo.Mode())
	}
	return true, RenameAndSync(f.Name(), path)
}
