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
				{"a", ""},
			},
			want: &Inventory{},
		},
		"java": {
			files: []fi{{"a.java", "a"}},
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Java", TotalBytes: 1},
				},
			},
		},
		"go": {
			files: []fi{{"a.go", "a"}},
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Go", TotalBytes: 1},
				},
			},
		},
		"java and go": {
			files: []fi{{"a.java", "aa"}, {"a.go", "a"}},
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Java", TotalBytes: 2},
					{Name: "Go", TotalBytes: 1},
				},
			},
		},
		"large": {
			files: []fi{
				{"a.java", "aaaaaaaaa"},
				{"b.java", "bbbbbbb"},
				{"a.go", "aaaaa"},
				{"b.go", "bbb"},
				{"c.txt", "ccccc"},
			},
			want: &Inventory{
				Languages: []*Lang{
					{Name: "Java", TotalBytes: 16},
					{Name: "Go", TotalBytes: 8},
					{Name: "Text", TotalBytes: 5},
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

	want := `{"Languages":[{"Name":"Go","TotalBytes":1505},{"Name":"Markdown","TotalBytes":38},{"Name":"YAML","TotalBytes":29},{"Name":"HTML","TotalBytes":28},{"Name":"Unix Assembly","TotalBytes":26},{"Name":"Protocol Buffer","TotalBytes":25},{"Name":"JavaScript","TotalBytes":16},{"Name":"CSS","TotalBytes":10},{"Name":"Perl","TotalBytes":9},{"Name":"JSON","TotalBytes":5},{"Name":"Shell","TotalBytes":4},{"Name":"Text","TotalBytes":3},{"Name":"INI","TotalBytes":2},{"Name":"SVG","TotalBytes":2},{"Name":"C","TotalBytes":1},{"Name":"Ignore List","TotalBytes":1},{"Name":"Python","TotalBytes":1},{"Name":"XML","TotalBytes":1}]}`
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
