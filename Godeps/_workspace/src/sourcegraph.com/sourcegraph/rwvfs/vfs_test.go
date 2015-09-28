package rwvfs_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"sourcegraph.com/sourcegraph/rwvfs"
	"sourcegraph.com/sourcegraph/rwvfs/testutil"

	"github.com/kr/fs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

func TestSub(t *testing.T) {
	m := rwvfs.Map(map[string]string{})
	sub := rwvfs.Sub(m, "/sub")

	err := sub.Mkdir("/")
	if err != nil {
		t.Fatal(err)
	}
	testutil.IsDir(t, "sub", m, "/sub")

	f, err := sub.Create("f1")
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	testutil.IsFile(t, "sub", m, "/sub/f1")

	f, err = sub.Create("/f2")
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	testutil.IsFile(t, "sub", m, "/sub/f2")

	err = sub.Mkdir("/d1")
	if err != nil {
		t.Fatal(err)
	}
	testutil.IsDir(t, "sub", m, "/sub/d1")

	err = sub.Mkdir("/d2")
	if err != nil {
		t.Fatal(err)
	}
	testutil.IsDir(t, "sub", m, "/sub/d2")
}

func TestRWVFS(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "rwvfs-test-")
	if err != nil {
		t.Fatal("TempDir", err)
	}
	defer os.RemoveAll(tmpdir)

	h := http.Handler(rwvfs.HTTPHandler(rwvfs.Map(map[string]string{}), nil))
	httpServer := httptest.NewServer(h)
	defer httpServer.Close()
	httpURL, err := url.Parse(httpServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		fs rwvfs.FileSystem
	}{
		{rwvfs.OS(tmpdir)},
		{rwvfs.Map(map[string]string{})},
		{rwvfs.Sub(rwvfs.Map(map[string]string{}), "/x")},
		{rwvfs.HTTP(httpURL, nil)},
		{rwvfs.Union(rwvfs.Map(map[string]string{}), rwvfs.Map(map[string]string{}))},
	}
	for _, test := range tests {
		testutil.Write(t, test.fs)
		testutil.Mkdir(t, test.fs)
		testutil.MkdirAll(t, test.fs)
		testutil.Glob(t, test.fs)
	}
}

func TestMap_MkdirAllWithRootNotExists(t *testing.T) {
	m := map[string]string{}
	fs := rwvfs.Sub(rwvfs.Map(m), "x")

	paths := []string{"a/b", "/c/d"}
	for _, path := range paths {
		if err := rwvfs.MkdirAll(fs, path); err != nil {
			t.Errorf("MkdirAll %q: %s", path, err)
		}
	}
}

func TestHTTP_BaseURL(t *testing.T) {
	m := map[string]string{"b/c": "c"}
	mapFS := rwvfs.Map(m)

	prefix := "/foo/bar/baz"

	h := http.Handler(http.StripPrefix(prefix, rwvfs.HTTPHandler(mapFS, nil)))
	httpServer := httptest.NewServer(h)
	defer httpServer.Close()
	httpURL, err := url.Parse(httpServer.URL + prefix)
	if err != nil {
		t.Fatal(err)
	}

	fs := rwvfs.HTTP(httpURL, nil)

	if err := rwvfs.MkdirAll(fs, "b"); err != nil {
		t.Errorf("MkdirAll %q: %s", "b", err)
	}

	fis, err := fs.ReadDir("b")
	if err != nil {
		t.Fatal(err)
	}
	if len(fis) != 1 {
		t.Errorf("got len(fis) == %d, want 1", len(fis))
	}
	if wantName := "c"; fis[0].Name() != wantName {
		t.Errorf("got name == %q, want %q", fis[0].Name(), wantName)
	}
}

func TestMap_Walk(t *testing.T) {
	m := map[string]string{"a": "a", "b/c": "c", "b/x/y/z": "z"}
	mapFS := rwvfs.Map(m)

	var names []string
	w := fs.WalkFS(".", rwvfs.Walkable(mapFS))
	for w.Step() {
		if err := w.Err(); err != nil {
			t.Fatalf("walk path %q: %s", w.Path(), err)
		}
		names = append(names, w.Path())
	}

	wantNames := []string{".", "a", "b", "b/c", "b/x", "b/x/y", "b/x/y/z"}
	sort.Strings(names)
	sort.Strings(wantNames)
	if !reflect.DeepEqual(names, wantNames) {
		t.Errorf("got entry names %v, want %v", names, wantNames)
	}
}

func TestMap_Walk2(t *testing.T) {
	m := map[string]string{"a/b/c/d": "a"}
	mapFS := rwvfs.Map(m)

	var names []string
	w := fs.WalkFS(".", rwvfs.Walkable(rwvfs.Sub(mapFS, "a/b")))
	for w.Step() {
		if err := w.Err(); err != nil {
			t.Fatalf("walk path %q: %s", w.Path(), err)
		}
		names = append(names, w.Path())
	}

	wantNames := []string{".", "c", "c/d"}
	sort.Strings(names)
	sort.Strings(wantNames)
	if !reflect.DeepEqual(names, wantNames) {
		t.Errorf("got entry names %v, want %v", names, wantNames)
	}
}

func TestReadOnly(t *testing.T) {
	m := map[string]string{"x": "y"}
	rfs := mapfs.New(m)
	wfs := rwvfs.ReadOnly(rfs)

	if _, err := rfs.Stat("/x"); err != nil {
		t.Error(err)
	}

	_, err := wfs.Create("/y")
	if want := (&os.PathError{"create", "/y", rwvfs.ErrReadOnly}); !reflect.DeepEqual(err, want) {
		t.Errorf("Create: got err %v, want %v", err, want)
	}

	err = wfs.Mkdir("/y")
	if want := (&os.PathError{"mkdir", "/y", rwvfs.ErrReadOnly}); !reflect.DeepEqual(err, want) {
		t.Errorf("Mkdir: got err %v, want %v", err, want)
	}

	err = wfs.Remove("/y")
	if want := (&os.PathError{"remove", "/y", rwvfs.ErrReadOnly}); !reflect.DeepEqual(err, want) {
		t.Errorf("Remove: got err %v, want %v", err, want)
	}

}

func TestOS_Symlink(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "rwvfs-test-")
	if err != nil {
		t.Fatal("TempDir", err)
	}
	defer os.RemoveAll(tmpdir)
	want := "hello"

	if err := ioutil.WriteFile(filepath.Join(tmpdir, "myfile"), []byte(want), 0600); err != nil {
		t.Fatal(err)
	}

	osfs := rwvfs.OS(tmpdir)
	if err := osfs.(rwvfs.LinkFS).Symlink("myfile", "mylink"); err != nil {
		t.Fatal(err)
	}
	got, err := ioutil.ReadFile(filepath.Join(tmpdir, "mylink"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("%s: ReadLink: got %q, want %q", osfs, string(got), want)
	}
}

func TestOS_Symlink_walkable(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "rwvfs-test-")
	if err != nil {
		t.Fatal("TempDir", err)
	}
	defer os.RemoveAll(tmpdir)
	want := "hello"

	if err := ioutil.WriteFile(filepath.Join(tmpdir, "myfile"), []byte(want), 0600); err != nil {
		t.Fatal(err)
	}

	osfs := rwvfs.OS(tmpdir)
	if err := rwvfs.Walkable(osfs).(rwvfs.LinkFS).Symlink("myfile", "mylink"); err != nil {
		t.Fatal(err)
	}
	got, err := ioutil.ReadFile(filepath.Join(tmpdir, "mylink"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("%s: ReadLink: got %q, want %q", osfs, string(got), want)
	}
}

func TestSub_Symlink(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "rwvfs-test-")
	if err != nil {
		t.Fatal("TempDir", err)
	}
	//defer os.RemoveAll(tmpdir)
	want := "hello"

	if err := os.Mkdir(filepath.Join(tmpdir, "mydir"), 0700); err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(filepath.Join(tmpdir, "mydir", "myfile"), []byte(want), 0600); err != nil {
		t.Fatal(err)
	}

	osfs := rwvfs.OS(tmpdir)
	sub := rwvfs.Sub(osfs, "mydir")
	if err := sub.(rwvfs.LinkFS).Symlink("myfile", "mylink"); err != nil {
		t.Fatal(err, osfs)
	}
	got, err := ioutil.ReadFile(filepath.Join(tmpdir, "mydir", "mylink"))
	if err != nil {
		t.Fatal(err, osfs, sub)
	}
	if string(got) != want {
		t.Errorf("%s: ReadLink: got %q, want %q", osfs, string(got), want)
	}
}

func TestOS_ReadLink(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "rwvfs-test-")
	if err != nil {
		t.Fatal("TempDir", err)
	}
	defer os.RemoveAll(tmpdir)

	if err := ioutil.WriteFile(filepath.Join(tmpdir, "myfile"), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(tmpdir, "myfile"), filepath.Join(tmpdir, "mylink")); err != nil {
		t.Fatal(err)
	}

	osfs := rwvfs.OS(tmpdir)
	dst, err := osfs.(rwvfs.LinkFS).ReadLink("mylink")
	if err != nil {
		t.Fatal(err)
	}
	if want := "myfile"; dst != want {
		t.Errorf("%s: ReadLink: got %q, want %q", osfs, dst, want)
	}
}

func TestOS_ReadLink_walkable(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "rwvfs-test-")
	if err != nil {
		t.Fatal("TempDir", err)
	}
	defer os.RemoveAll(tmpdir)

	if err := ioutil.WriteFile(filepath.Join(tmpdir, "myfile"), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(tmpdir, "myfile"), filepath.Join(tmpdir, "mylink")); err != nil {
		t.Fatal(err)
	}

	osfs := rwvfs.OS(tmpdir)
	dst, err := rwvfs.Walkable(osfs).(rwvfs.LinkFS).ReadLink("mylink")
	if err != nil {
		t.Fatal(err)
	}
	if want := "myfile"; dst != want {
		t.Errorf("%s: ReadLink: got %q, want %q", osfs, dst, want)
	}
}

func TestSub_ReadLink(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "rwvfs-test-")
	if err != nil {
		t.Fatal("TempDir", err)
	}
	defer os.RemoveAll(tmpdir)

	if err := os.Mkdir(filepath.Join(tmpdir, "mydir"), 0700); err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(filepath.Join(tmpdir, "mydir", "myfile"), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(tmpdir, "mydir", "myfile"), filepath.Join(tmpdir, "mydir", "mylink")); err != nil {
		t.Fatal(err)
	}

	osfs := rwvfs.OS(tmpdir)
	sub := rwvfs.Sub(osfs, "mydir")
	dst, err := sub.(rwvfs.LinkFS).ReadLink("mylink")
	if err != nil {
		t.Fatal(err)
	}
	if want := "myfile"; dst != want {
		t.Errorf("%s: ReadLink: got %q, want %q", osfs, dst, want)
	}
}

func TestOS_ReadLink_ErrOutsideRoot(t *testing.T) {
	tmpdir1, err := ioutil.TempDir("", "rwvfs-test-")
	if err != nil {
		t.Fatal("TempDir", err)
	}
	defer os.RemoveAll(tmpdir1)

	tmpdir2, err := ioutil.TempDir("", "rwvfs-test-")
	if err != nil {
		t.Fatal("TempDir", err)
	}
	defer os.RemoveAll(tmpdir2)

	if err := ioutil.WriteFile(filepath.Join(tmpdir1, "myfile"), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(tmpdir1, "myfile"), filepath.Join(tmpdir2, "mylink")); err != nil {
		t.Fatal(err)
	}

	osfs := rwvfs.OS(tmpdir2)
	dst, err := osfs.(rwvfs.LinkFS).ReadLink("mylink")
	if want := rwvfs.ErrOutsideRoot; err != want {
		t.Fatalf("%s: ReadLink: got err %v, want %v", osfs, err, want)
	}
	if want := filepath.Join(tmpdir1, "myfile"); dst != want {
		t.Errorf("%s: ReadLink: got %q, want %q", osfs, dst, want)
	}
}

func TestUnion(t *testing.T) {
	m1 := map[string]string{"foo/file": ""}
	m2 := map[string]string{"bar/file": ""}
	u := rwvfs.Union(rwvfs.Map(m1), rwvfs.Map(m2))

	infos, err := u.ReadDir("/")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(infos), 2; got != want {
		t.Errorf(`ReadDir: got %d, want %d`, got, want)
	}

	testCreate(t, u, "test", m1)
	testCreate(t, u, "foo/test", m1)
	testCreate(t, u, "bar/test", m2)
}

func TestEmptyUnion(t *testing.T) {
	u := rwvfs.Union()
	infos, err := u.ReadDir("/")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(infos), 0; got != want {
		t.Errorf(`ReadDir: got %d, want %d`, got, want)
	}

	_, err = u.Create("/foo")
	if err == nil {
		t.Error("Create: read-only error expected")
	}
}

func testCreate(t *testing.T, u rwvfs.FileSystem, path string, m map[string]string) {
	w, err := u.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("test"))
	w.Close()

	if got, want := m[path], "test"; got != want {
		t.Errorf(`Create: got %v, want %v`, got, want)
	}
}
