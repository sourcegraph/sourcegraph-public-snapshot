package util

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
// - rootPath refers to temporary root directory the filePath is in
// (everything under the rootPath (including rootPath) should be removed when no longer needed)
// - error indicates possible error
func ScriptFile(prefix string) (filePath string, rootPath string, err error) {
	var suffix string
	if runtime.GOOS == "windows" {
		suffix = ".bat"
	}

	tempDir, err := ioutil.TempDir("", prefix)
	if err != nil {
		return "", "", err
	}
	return filepath.Join(tempDir, prefix+suffix), tempDir, nil
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_959(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
