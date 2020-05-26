package testutil

import (
	"github.com/google/go-cmp/cmp"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func Diff(b1, b2 string) (string, error) {
	f1, err := ioutil.TempFile("", "diff_test")
	if err != nil {
		return "", err
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("", "diff_test")
	if err != nil {
		return "", err
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	_, err = f1.WriteString(b1)
	if err != nil {
		return "", err
	}
	_, err = f2.WriteString(b2)
	if err != nil {
		return "", err
	}

	data, err := exec.Command("diff", "-u", "--label=want", f1.Name(), "--label=got", f2.Name()).CombinedOutput()
	if len(data) > 0 {
		err = nil
	}
	return string(data), err
}

// DeepCompare compares x and y using cmp.Diff and raises a fatal error
// on t if there are any differences.
func DeepCompare(t *testing.T, x, y interface{}, opts ...cmp.Option) {
	t.Helper()
	if diff := cmp.Diff(x, y, opts...); diff != "" {
		t.Fatal(diff)
	}
}
