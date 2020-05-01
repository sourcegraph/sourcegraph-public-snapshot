package janitor

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func withRoot(t *testing.T, testFunc func(bundleDir string)) {
	bundleDir, err := ioutil.TempDir("", "precise-code-intel-bundle-manager-")
	if err != nil {
		t.Fatalf("unexpected error creating test directory: %s", err)
	}
	defer os.RemoveAll(bundleDir)

	for _, dir := range []string{"", "uploads", "dbs"} {
		path := filepath.Join(bundleDir, dir)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			t.Fatalf("unexpected error creating test directory: %s", err)
		}
	}

	testFunc(bundleDir)
}

func makeFile(path string, mtimes time.Time) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	file.Close()

	return os.Chtimes(path, mtimes, mtimes)
}

func makeFileWithSize(path string, size int) error {
	return ioutil.WriteFile(path, make([]byte, size), 0644)
}

func getFilenames(path string) ([]string, error) {
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, info := range infos {
		names = append(names, info.Name())
	}
	sort.Strings(names)

	return names, nil
}
