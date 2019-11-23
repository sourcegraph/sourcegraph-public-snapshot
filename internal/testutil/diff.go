package testutil

import (
	"io/ioutil"
	"os"
	"os/exec"
)

func Diff(b1, b2 string) (string, error) {
	f1, err := ioutil.TempFile("", "test_diff")
	if err != nil {
		return "", err
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("", "test_diff")
	if err != nil {
		return "", err
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	if _, err := f1.WriteString(b1); err != nil {
		return "", err
	}
	if _, err := f2.WriteString(b2); err != nil {
		return "", err
	}

	data, err := exec.Command("diff", "-u", "--label=want", f1.Name(), "--label=got", f2.Name()).CombinedOutput()
	if len(data) > 0 {
		err = nil
	}
	return string(data), err
}
