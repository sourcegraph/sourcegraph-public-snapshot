package shared

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNginx(t *testing.T) {
	read := func(path string) []byte {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}

	dir := t.TempDir()

	path, err := nginxWriteFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "nginx.conf" {
		t.Fatalf("unexpected nginx.conf path: %s", path)
	}

	count := 0
	err = filepath.Walk("assets", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.Contains(path, "nginx") {
			return nil
		}

		path, err = filepath.Rel("assets", path)
		if err != nil {
			t.Fatal(err)
		}

		count++
		t.Log(path)
		want := read(filepath.Join("assets", path))
		got := read(filepath.Join(dir, path))
		if !bytes.Equal(want, got) {
			t.Fatalf("%s has different contents", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count < 2 {
		t.Fatal("did not find enough nginx configurations")
	}
}
