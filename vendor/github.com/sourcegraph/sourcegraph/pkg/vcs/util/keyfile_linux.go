package util

import (
	"fmt"
	"os"
	"syscall"
)

func writeKeyTempFile0(tmpfile *os.File, keyData []byte) (filename string, err error) {
	if err := syscall.Unlink(tmpfile.Name()); err != nil {
		return "", err
	}
	if _, err := tmpfile.Write(keyData); err != nil {
		return "", err
	}
	return fmt.Sprintf("/proc/self/fd/%d", tmpfile.Fd()), nil
}
