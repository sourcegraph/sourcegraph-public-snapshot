package util

import (
	"os"
	"path/filepath"
	"runtime"
)

// Constructs platform-specific temporary script file with a given prefix
// On Windows such a file must have .bat extension
// Returns triplet where
// - filePath is a location of file
// - rootPath refers to temporary root directory the filePath is in
// (everything under the rootPath (including rootPath) should be removed when no longer needed)
// - error indicates possible error
func ScriptFile(prefix string) (filePath string, rootPath string, err error) {
	var suffix string
	if runtime.GOOS == "windows" {
		suffix = ".bat"
	}

	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return "", "", err
	}
	return filepath.Join(tempDir, prefix+suffix), tempDir, nil
}

// Wrapper around os.WriteFile that updates permissions regardless if file existed before
func WriteFileWithPermissions(file string, content []byte, perm os.FileMode) error {
	err := os.WriteFile(file, content, perm)
	if err != nil {
		return err
	}
	// os.WriteFile applies permissions only for files that weren't exist
	return os.Chmod(file, perm)
}
