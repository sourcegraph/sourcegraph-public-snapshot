package golang

import (
	"testing"

	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

func TestNamespaceFS(t *testing.T) {
	mfs := mapfs.New(map[string]string{"/a/b.txt": "c"})
	ns := vfs.NewNameSpace()
	ns.Bind("/x/y", mfs, "/", vfs.BindBefore)

	fs := namespaceFS{ns}

	statTests := []struct {
		path      string
		wantIsDir bool
	}{
		{"/", true},
		{"/x", true},
		{"/x/y", true},
		{"/x/y/a", true},
	}
	for _, test := range statTests {
		fi, err := fs.Stat(test.path)
		if err != nil {
			t.Errorf("Stat(%q): %s", test.path, err)
			continue
		}
		if fi.Mode().IsDir() != test.wantIsDir {
			t.Errorf("Stat(%q): got IsDir %v, want %v", test.path, fi.Mode().IsDir(), test.wantIsDir)
		}
	}
}
