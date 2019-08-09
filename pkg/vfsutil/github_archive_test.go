package vfsutil

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestGitHubRepoVFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skip network-intensive test")
	}

	// Ensure fetch logic works
	cleanup := useEmptyArchiveCacheDir()
	defer cleanup()

	// Any public repo will work.
	fs, err := NewGitHubRepoVFS("github.com/gorilla/schema", "0164a00ab4cd01d814d8cd5bf63fd9fcea30e23b")
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Close()
	want := map[string]string{
		"/LICENSE":         "...",
		"/README.md":       "schema...",
		"/cache.go":        "// Copyright...",
		"/converter.go":    "// Copyright...",
		"/decoder.go":      "// Copyright...",
		"/decoder_test.go": "// Copyright...",
		"/doc.go":          "// Copyright...",
		"/.travis.yml":     "...",
	}

	testVFS(t, fs, want)
}

func useEmptyArchiveCacheDir() func() {
	d, err := ioutil.TempDir("", "vfsutil_test")
	if err != nil {
		panic(err)
	}
	orig := ArchiveCacheDir
	ArchiveCacheDir = d
	return func() {
		os.RemoveAll(d)
		ArchiveCacheDir = orig
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_969(size int) error {
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
