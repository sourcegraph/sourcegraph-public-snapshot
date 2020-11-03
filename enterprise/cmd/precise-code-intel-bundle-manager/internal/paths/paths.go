package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

const uploadDir = "uploads"
const uploadPartsDir = "upload-parts"
const dbsDir = "dbs"
const dbPartsDir = "db-parts"
const dbBackupsDir = "db-backups"
const migrationMarkersDir = "migration-markers"

// PrepDirectories creates the root directories within the given bundle dir.
func PrepDirectories(bundleDir string) error {
	rootDirs := []string{
		uploadDir,
		uploadPartsDir,
		dbsDir,
		dbPartsDir,
		dbBackupsDir,
		migrationMarkersDir,
	}

	for _, dir := range rootDirs {
		if err := os.MkdirAll(filepath.Join(bundleDir, dir), os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

// UploadsDir returns the path of the directory containing upload files.
func UploadsDir(bundleDir string) string {
	return filepath.Join(bundleDir, uploadDir)
}

// UploadFilename returns the path of the upload with the given identifier.
func UploadFilename(bundleDir string, id int64) string {
	return filepath.Join(bundleDir, uploadDir, fmt.Sprintf("%d.gz", id))
}

// UploadPartsDir returns the path of the directory containing upload part files.
func UploadPartsDir(bundleDir string) string {
	return filepath.Join(bundleDir, uploadPartsDir)
}

// UploadPartFilename returns the path of the upload with the given identifier and part index.
func UploadPartFilename(bundleDir string, id, index int64) string {
	return filepath.Join(bundleDir, uploadPartsDir, fmt.Sprintf("%d.%d.gz", id, index))
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
