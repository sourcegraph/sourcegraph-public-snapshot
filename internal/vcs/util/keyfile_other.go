// +build !linux

package util

import "os"

func writeKeyTempFile0(tmpfile *os.File, keyData []byte) (filename string, err error) {
	if _, err := tmpfile.Write(keyData); err != nil {
		return "", err
	}
	return tmpfile.Name(), nil
}
