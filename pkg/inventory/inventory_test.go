package inventory

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestScan(t *testing.T) {
	tests := map[string]struct {
		files   []fi
		want    *Inventory
		wantErr error
	}{
		"empty file": {
			files: []fi{
				fi{"a", ""},
			},
			want: &Inventory{},
		},
		"excludes": {
			files: []fi{
				fi{"a.min.js", "a"},
				fi{"node_modules/a/b.js", "a"},
				fi{"Godeps/_workspace/src/a/b.go", "a"},
			},
			want: &Inventory{},
		},
		"java": {
			files: []fi{fi{"a.java", "a"}},
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Java", TotalBytes: 1, Type: "programming"},
				},
			},
		},
		"go": {
			files: []fi{fi{"a.go", "a"}},
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Go", TotalBytes: 1, Type: "programming"},
				},
			},
		},
		"java and go": {
			files: []fi{fi{"a.java", "aa"}, fi{"a.go", "a"}},
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Java", TotalBytes: 2, Type: "programming"},
					{Name: "Go", TotalBytes: 1, Type: "programming"},
				},
			},
		},
		"large": {
			files: []fi{
				fi{"a.java", "aaaaaaaaa"},
				fi{"b.java", "bbbbbbb"},
				fi{"a.go", "aaaaa"},
				fi{"b.go", "bbb"},
				fi{"c.txt", "ccccc"},
			},
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
		var fi []os.FileInfo
		for _, file := range test.files {
			fi = append(fi, file)
		}
		inv, err := Get(context.Background(), fi)
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

type fi struct {
	Path     string
	Contents string
}

func (f fi) Name() string {
	return f.Path
}
func (f fi) Size() int64 {
	return int64(len(f.Contents))
}
func (f fi) IsDir() bool {
	return false
}
func (f fi) Mode() os.FileMode {
	return os.FileMode(0)
}
func (f fi) ModTime() time.Time {
	return time.Now()
}
func (f fi) Sys() interface{} {
	return interface{}(nil)
}
