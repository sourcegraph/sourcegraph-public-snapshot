package inventory

import (
	"errors"
	pathpkg "path"
	"reflect"
	"testing"

	"github.com/kr/fs"
	"golang.org/x/net/context"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

func TestScan(t *testing.T) {
	tests := map[string]struct {
		fs      fs.FileSystem
		want    *Inventory
		wantErr error
	}{
		"no files": {
			fs:      walkableFileSystem{mapfs.New(map[string]string{})},
			wantErr: errors.New("file does not exist"),
		},
		"empty file": {
			fs:   walkableFileSystem{mapfs.New(map[string]string{"a": ""})},
			want: &Inventory{},
		},
		"excludes": {
			fs: walkableFileSystem{mapfs.New(map[string]string{
				"a.min.js":                     "a",
				"node_modules/a/b.js":          "a",
				"Godeps/_workspace/src/a/b.go": "a",
			})},
			want: &Inventory{},
		},
		"java": {
			fs: walkableFileSystem{mapfs.New(map[string]string{"a.java": "a"})},
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Java", TotalBytes: 1, Type: "programming"},
				},
			},
		},
		"go": {
			fs: walkableFileSystem{mapfs.New(map[string]string{"a.go": "a"})},
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Go", TotalBytes: 1, Type: "programming"},
				},
			},
		},
		"java and go": {
			fs: walkableFileSystem{mapfs.New(map[string]string{"a.java": "aa", "a.go": "a"})},
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Java", TotalBytes: 2, Type: "programming"},
					{Name: "Go", TotalBytes: 1, Type: "programming"},
				},
			},
		},
		"large": {
			fs: walkableFileSystem{mapfs.New(map[string]string{
				"a.java": "aaaaaaaaa",
				"b.java": "bbbbbbb",
				"a.go":   "aaaaa",
				"b.go":   "bbb",
				"c.txt":  "ccccc",
			})},
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Java", TotalBytes: 16, Type: "programming"},
					{Name: "Go", TotalBytes: 8, Type: "programming"},
					{Name: "Text", TotalBytes: 5, Type: "prose"},
				},
			},
		},
	}
	for label, test := range tests {
		inv, err := Scan(context.Background(), test.fs)
		if err != nil && (test.wantErr == nil || err.Error() != test.wantErr.Error()) {
			t.Errorf("%s: Scan: %s (want error %v)", label, err, test.wantErr)
			continue
		}
		if test.wantErr != nil && err == nil {
			t.Errorf("%s: Scan: got error == nil, want error %v", label, test.wantErr)
			continue
		}
		if !reflect.DeepEqual(inv, test.want) {
			t.Errorf("%s: got %+v, want %+v", label, inv, test.want)
			continue
		}
	}
}

type walkableFileSystem struct{ vfs.FileSystem }

func (walkableFileSystem) Join(path ...string) string { return pathpkg.Join(path...) }
