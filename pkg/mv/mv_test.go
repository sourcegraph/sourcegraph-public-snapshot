package mv

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestMove(t *testing.T) {
	dir, err := ioutil.TempDir("", "mvtest")
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	err = os.Mkdir(src, 0700)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(src, "test"), []byte("Hello World!"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	err = move(src, dst)
	if err != nil {
		t.Fatal(err)
	}

	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(fis) != 1 || fis[0].Name() != "dst" {
		t.Errorf("move either did not delete src or did not create dst: %#v", fis)
	}

	fis, err = ioutil.ReadDir(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(fis) != 1 || fis[0].Name() != "test" {
		t.Errorf("dir/test missing: %#v", fis)
	}

	b, err := ioutil.ReadFile(filepath.Join(dst, "test"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "Hello World!" {
		t.Errorf("Unexpected content in dst/test: %v", string(b))
	}
}
