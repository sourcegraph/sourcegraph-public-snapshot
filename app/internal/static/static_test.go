package static

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
)

func TestStatic_file(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "static")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if err := ioutil.WriteFile(filepath.Join(tmpDir, "foo.css"), []byte("bar"), 0600); err != nil {
		t.Fatal(err)
	}

	Flags.Dir = tmpDir
	defer func() {
		Flags.Dir = ""
	}()

	reuse = nil
	c, _ := apptest.New()

	resp, err := c.GetNoFollowRedirects("/foo.css")
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("got HTTP status %d, want %d", resp.Status, want)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if want := "bar"; string(body) != want {
		t.Errorf("got body %q, want %q", body, want)
	}
}

func TestStatic_tmpl(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "static")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if err := ioutil.WriteFile(filepath.Join(tmpDir, "foo.tmpl"), []byte("bar at {{.CurrentURL.Host}}"), 0600); err != nil {
		t.Fatal(err)
	}

	Flags.Dir = tmpDir
	defer func() {
		Flags.Dir = ""
	}()

	reuse = nil
	c, _ := apptest.New()

	resp, err := c.GetNoFollowRedirects("/foo")
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("got HTTP status %d, want %d", resp.StatusCode, want)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if want := "bar at example.com"; string(body) != want {
		t.Errorf("got body %q, want %q", body, want)
	}
}
