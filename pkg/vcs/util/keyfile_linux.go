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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_956(size int) error {
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
