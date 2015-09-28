package util

import (
	"crypto/sha256"
	"encoding/base64"
	"io/ioutil"
	"os"
)

// WriteKeyTempFile writes data to a temp file whose filename contains
// some function of namePrefix. On Linux, the temp file is unlinked
// and the filename by which to access it is /proc/self/fd/N, where N
// is the fd of the file. The caller should call the Remove method on
// the returned File when done using it.
func WriteKeyTempFile(namePrefix string, keyData []byte) (filename string, tmp *os.File, err error) {
	hasher := sha256.New()
	hasher.Write([]byte(namePrefix))
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	tmpfile, err := ioutil.TempFile("", "go-vcs-"+hash+"-")
	if err != nil {
		return "", nil, err
	}

	filename, err = writeKeyTempFile0(tmpfile, keyData)
	return filename, tmpfile, err
}
