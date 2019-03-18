package inventory

import (
	"bufio"
	"context"
	"encoding/json"
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

func BenchmarkGet(b *testing.B) {
	files, err := readFileTree("prom-repo-tree.txt")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err = Get(context.Background(), files)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestGetGolden(t *testing.T) {
	mustMarshal := func(v interface{}) string {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	files, err := readFileTree("prom-repo-tree.txt")
	if err != nil {
		t.Fatal(err)
	}

	want := `{"Languages":[{"Name":"Go","TotalBytes":1505,"Type":"programming"},{"Name":"Markdown","TotalBytes":38,"Type":"prose"},{"Name":"YAML","TotalBytes":29,"Type":"data"},{"Name":"HTML","TotalBytes":28,"Type":"markup"},{"Name":"Unix Assembly","TotalBytes":26,"Type":"programming"},{"Name":"Protocol Buffer","TotalBytes":25,"Type":"data"},{"Name":"JavaScript","TotalBytes":16,"Type":"programming"},{"Name":"CSS","TotalBytes":10,"Type":"markup"},{"Name":"Perl","TotalBytes":9,"Type":"programming"},{"Name":"JSON","TotalBytes":5,"Type":"data"},{"Name":"Text","TotalBytes":4,"Type":"prose"},{"Name":"Shell","TotalBytes":4,"Type":"programming"},{"Name":"SVG","TotalBytes":2,"Type":"data"},{"Name":"INI","TotalBytes":2,"Type":"data"},{"Name":"XML","TotalBytes":1,"Type":"data"},{"Name":"Python","TotalBytes":1,"Type":"programming"},{"Name":"Makefile","TotalBytes":1,"Type":"programming"},{"Name":"Dockerfile","TotalBytes":1,"Type":"data"},{"Name":"C","TotalBytes":1,"Type":"programming"}]}`
	got, err := Get(context.Background(), files)
	if err != nil {
		t.Fatal(err)
	}
	if mustMarshal(got) != want {
		t.Errorf("did not match golden\ngot:  %s\nwant: %s", mustMarshal(got), want)
	}
}

func readFileTree(name string) ([]os.FileInfo, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var files []os.FileInfo
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		files = append(files, fi{scanner.Text(), "a"})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return files, nil
}
