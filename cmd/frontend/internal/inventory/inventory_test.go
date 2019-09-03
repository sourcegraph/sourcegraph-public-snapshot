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

func TestGet_noReadFile(t *testing.T) {
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
		inv, err := Get(context.Background(), fi, nil)
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

func TestGet_readFile(t *testing.T) {
	files := []os.FileInfo{
		fi{"a.java", "aaaaaaaaa"},
		fi{"b.md", "# Hello"},

		// The .m extension is used by many languages, but this code is obviously Objective-C. This
		// test checks that this file is detected correctly as Objective-C.
		fi{"c.m", "@interface X:NSObject { double x; } @property(nonatomic, readwrite) double foo;"},
	}
	inv, err := Get(context.Background(), files, func(_ context.Context, path string, maxFileBytes int64) ([]byte, error) {
		for _, f := range files {
			if f.Name() == path {
				return []byte(f.(fi).Contents), nil
			}
		}
		panic("no file: " + path)
	})
	if err != nil {
		t.Fatal(err)
	}

	want := &Inventory{
		Languages: []*Lang{
			{
				Name:       "Objective-C",
				TotalBytes: 79,
			},
			{
				Name:       "Java",
				TotalBytes: 9,
			},
			{
				Name:       "Markdown",
				TotalBytes: 7,
			},
		},
	}
	if !reflect.DeepEqual(inv, want) {
		t.Errorf("got  %+v\nwant %+v", mustMarshal(inv), mustMarshal(want))
	}
}

func BenchmarkGet(b *testing.B) {
	files, err := readFileTree("prom-repo-tree.txt")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err = Get(context.Background(), files, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestGetGolden(t *testing.T) {
	files, err := readFileTree("prom-repo-tree.txt")
	if err != nil {
		t.Fatal(err)
	}

	want := `{"Languages":[{"Name":"Go","TotalBytes":140},{"Name":"HTML","TotalBytes":28},{"Name":"Markdown","TotalBytes":10},{"Name":"YAML","TotalBytes":8},{"Name":"CSS","TotalBytes":4},{"Name":"JavaScript","TotalBytes":3},{"Name":"JSON","TotalBytes":1},{"Name":"Protocol Buffer","TotalBytes":1},{"Name":"SVG","TotalBytes":1},{"Name":"Shell","TotalBytes":1},{"Name":"XML","TotalBytes":1}]}`
	got, err := Get(context.Background(), files, nil)
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

func mustMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
