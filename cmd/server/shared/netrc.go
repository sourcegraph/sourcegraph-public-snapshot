package shared

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// copyNetrc will copy the file at /etc/sourcegraph/netrc to /etc/netrc for
// authenticated HTTP(S) cloning.
func copyNetrc() error {
	src := filepath.Join(os.Getenv("CONFIG_DIR"), "netrc")
	dst := os.ExpandEnv("$HOME/.netrc")

	data, err := ioutil.ReadFile(src)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return ioutil.WriteFile(dst, data, 0600)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_539(size int) error {
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
