package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

// PrepDirectories
func PrepDirectories(bundleDir string) error {
	for _, dir := range []string{UploadsDir(bundleDir), DBsDir(bundleDir)} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

// UploadsDir returns the path of the uploads directory.
func UploadsDir(bundleDir string) string {
	return filepath.Join(bundleDir, "uploads")
}

// DBsDir returns the path of the dbs directory.
func DBsDir(bundleDir string) string {
	return filepath.Join(bundleDir, "dbs")
}

// UploadFilename returns the path fo the upload with the given identifier.
func UploadFilename(bundleDir string, id int64) string {
	return filepath.Join(bundleDir, "uploads", fmt.Sprintf("%d.lsif.gz", id))
}

func UploadPartFilename(bundleDir string, id, index int64) string {
	return filepath.Join(bundleDir, "uploads", fmt.Sprintf("%d.%d.lsif.gz", id, index))
}

// DBFilename returns the path fo the database with the given identifier.
func DBFilename(bundleDir string, id int64) string {
	return filepath.Join(bundleDir, "dbs", fmt.Sprintf("%d.lsif.db", id))
}

// PathExists returns (true, nil) if the specified path exists, or (false, error) if an error
// occurred (such as not having permission to read the path).
func PathExists(filename string) (bool, error) {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
