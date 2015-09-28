package makex

import (
	"io/ioutil"
	"os"
	"os/exec"
)

// External runs makefile using the system `make` tool, optionally passing extra
// args.
func External(dir string, makefile []byte, args []string) ([]byte, error) {
	tmpFile, err := ioutil.TempFile("", "sg-makefile")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	err = ioutil.WriteFile(tmpFile.Name(), makefile, 0600)
	if err != nil {
		return nil, err
	}

	args = append(args, "-f", tmpFile.Name())
	mk := exec.Command("make", args...)
	mk.Dir = dir
	return mk.CombinedOutput()
}
