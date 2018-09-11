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
