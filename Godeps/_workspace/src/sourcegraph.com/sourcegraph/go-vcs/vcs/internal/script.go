package internal

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

var gen rand.Source

func init() {
	gen = rand.NewSource(time.Now().UnixNano())
}

// Constructs platform-specific temporary script file with a given prefix
// On Windows such a file must have .bat extension
func ScriptFile(prefix string) (string, error) {

	if runtime.GOOS == "windows" {
		for {
			tempFile := filepath.Join(os.TempDir(), prefix+strconv.FormatInt(gen.Int63(), 36)+".bat")
			_, err := os.Stat(tempFile)
			if err != nil {
				if os.IsNotExist(err) {
					return filepath.ToSlash(tempFile), nil
				} else {
					return "", err
				}
			}
		}
	} else {
		tf, err := ioutil.TempFile("", prefix)
		if err != nil {
			return "", err
		}
		tf.Close()
		return filepath.ToSlash(tf.Name()), nil
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
