package inventory

import (
	"errors"
	"path"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vfsutil"

	"context"

	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

func TestScan(t *testing.T) {
	tests := map[string]struct {
		fs      vfs.FileSystem
		want    *Inventory
		wantErr error
	}{
		"no files": {
			fs:      mapfs.New(map[string]string{}),
			wantErr: errors.New("file does not exist"),
		},
		"empty file": {
			fs:   mapfs.New(map[string]string{"a": ""}),
			want: &Inventory{},
		},
		"excludes": {
			fs: mapfs.New(map[string]string{
				"a.min.js":                     "a",
				"node_modules/a/b.js":          "a",
				"Godeps/_workspace/src/a/b.go": "a",
			}),
			want: &Inventory{},
		},
		"java": {
			fs: mapfs.New(map[string]string{"a.java": "a"}),
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Java", TotalBytes: 1, Type: "programming"},
				},
			},
		},
		"go": {
			fs: mapfs.New(map[string]string{"a.go": "a"}),
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Go", TotalBytes: 1, Type: "programming"},
				},
			},
		},
		"java and go": {
			fs: mapfs.New(map[string]string{"a.java": "aa", "a.go": "a"}),
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Java", TotalBytes: 2, Type: "programming"},
					{Name: "Go", TotalBytes: 1, Type: "programming"},
				},
			},
		},
		"large": {
			fs: mapfs.New(map[string]string{
				"a.java": "aaaaaaaaa",
				"b.java": "bbbbbbb",
				"a.go":   "aaaaa",
				"b.go":   "bbb",
				"c.txt":  "ccccc",
			}),
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
		inv, err := Scan(context.Background(), vfsutil.Walkable(test.fs, path.Join))
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
