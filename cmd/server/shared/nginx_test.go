package shared

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNginx(t *testing.T) {
	read := func(path string) []byte {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}

	dir, err := ioutil.TempDir("", "nginx_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path, err := nginxWriteFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "nginx.conf" {
		t.Fatalf("unexpected nginx.conf path: %s", path)
	}

	count := 0
	err = filepath.Walk("assets", func(path string, info os.FileInfo, err error) error {
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_541(size int) error {
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
