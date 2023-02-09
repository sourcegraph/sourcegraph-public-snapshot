package backend

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestAskCody(t *testing.T) {
	fmt.Println(os.Getwd())
	path := os.Getenv("FILE_PATH")
	if path == "" {
		t.Fatalf("Please set FILE_PATH env var")
	}
	var secFiles []string
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if len(secFiles) > 3 { // We just need a few for now
			return nil
		}
		fileContents, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "Could not read %q", path)
		}
		fmt.Println("PATH", path)
		hit, err := CheckLabel(string(fileContents), "security")
		if hit != nil {
			secFiles = append(secFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Errorf("Files that require security expertise: %v", secFiles)
}
