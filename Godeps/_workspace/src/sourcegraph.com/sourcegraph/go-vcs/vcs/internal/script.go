package internal

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

// Constructs platform-specific temporary script file with a given prefix
// On Windows such a file must have .bat extension
// Returns triplet where
// - filePath is a location of file
// - rootPath may refer to temporary root directory
// (everything under the rootPath (including rootPath) should be removed when no longer needed)
// rootPath makes sense on Windows only where location of script file is TEMP_DIR()/RANDOM_DIR()/FILE.bat
// - error indicates possible error
func ScriptFile(prefix string) (filePath string, rootPath string, err error) {

	if runtime.GOOS == "windows" {
		// making unique temporary directory and file inside it
		tempDir, err := ioutil.TempDir("", prefix)
		if err != nil {
			return "", "", err
		}
		return filepath.Join(tempDir, prefix+".bat"), tempDir, nil
	} else {
		tf, err := ioutil.TempFile("", prefix)
		if err != nil {
			return "", "", err
		}
		tf.Close()
		return filepath.ToSlash(tf.Name()), "", nil
	}
}

// Wrapper around ioutil.WriteFile that updates permissions regardless if file existed before
func WriteFileWithPermissions(file string, content []byte, perm os.FileMode) error {
	err := ioutil.WriteFile(file, content, perm)
	if err != nil {
		return err
	}
	// ioutil.WriteFile applies permissions only for files that weren't exist
	return os.Chmod(file, perm)
}
