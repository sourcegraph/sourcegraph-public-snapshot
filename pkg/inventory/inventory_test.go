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
		"java": {
			fs: walkableFileSystem{mapfs.New(map[string]string{"a.java": "a"})},
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Java", TotalBytes: 1},
				},
			},
		},
		"go": {
			fs: walkableFileSystem{mapfs.New(map[string]string{"a.go": "a"})},
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Go", TotalBytes: 1},
				},
			},
		},
		"java and go": {
			fs: walkableFileSystem{mapfs.New(map[string]string{"a.java": "aa", "a.go": "a"})},
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Java", TotalBytes: 2},
					{Name: "Go", TotalBytes: 1},
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
					{Name: "Java", TotalBytes: 16},
					{Name: "Go", TotalBytes: 8},
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
