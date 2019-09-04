package inventory

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/src-d/enry/v2"
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

func newReadFile(files []os.FileInfo) func(_ context.Context, path string, maxFileBytes int64) ([]byte, error) {
	m := make(map[string][]byte, len(files))
	for _, f := range files {
		m[f.Name()] = []byte(f.(fi).Contents)
	}
	return func(_ context.Context, path string, maxFileBytes int64) ([]byte, error) {
		data, ok := m[path]
		if !ok {
			return nil, fmt.Errorf("no file: %s", path)
		}
		return data, nil
	}
}

func TestGet_readFile(t *testing.T) {
	files := []os.FileInfo{
		fi{"a.java", "aaaaaaaaa"},
		fi{"b.md", "# Hello"},

		// The .m extension is used by many languages, but this code is obviously Objective-C. This
		// test checks that this file is detected correctly as Objective-C.
		fi{"c.m", "@interface X:NSObject { double x; } @property(nonatomic, readwrite) double foo;"},
	}
	inv, err := Get(context.Background(), files, newReadFile(files))
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

func TestOverrideTSXAndJSX(t *testing.T) {
	// Ensure that .tsx and .jsx are considered as valid extensions for TypeScript and JavaScript,
	// respectively.
	files := []os.FileInfo{
		fi{"a.tsx", "xx"},
		fi{"b.jsx", "x"},
	}
	inv, err := Get(context.Background(), files, newReadFile(files))
	if err != nil {
		t.Fatal(err)
	}

	want := &Inventory{
		Languages: []*Lang{
			{
				Name:       "TypeScript",
				TotalBytes: 2,
			},
			{
				Name:       "JavaScript",
				TotalBytes: 1,
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
	b.Logf("Calling Get on %d files.", len(files))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err = Get(context.Background(), files, newReadFile(files))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIsVendor(b *testing.B) {
	files, err := readFileTree("prom-repo-tree.txt")
	if err != nil {
		b.Fatal(err)
	}
	b.Logf("Calling IsVendor on %d files.", len(files))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, f := range files {
			_ = enry.IsVendor(f.Name())
		}
	}
}

func TestGetGolden(t *testing.T) {
	files, err := readFileTree("prom-repo-tree.txt")
	if err != nil {
		t.Fatal(err)
	}

	want := `{"Languages":[{"Name":"Go","TotalBytes":14980},{"Name":"Limbo","TotalBytes":11178},{"Name":"HTML","TotalBytes":2044},{"Name":"JavaScript","TotalBytes":273},{"Name":"CSS"},{"Name":"JSON"},{"Name":"Markdown"},{"Name":"Protocol Buffer"},{"Name":"SVG"},{"Name":"Shell"},{"Name":"XML"},{"Name":"YAML"}]}`
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
		path := scanner.Text()
		files = append(files, fi{path, fakeContents(path)})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

func fakeContents(path string) string {
	switch filepath.Ext(path) {
	case ".html":
		return `<html><head><title>hello</title></head><body><h1>hello</h1></body></html>`
	case ".go":
		return `package foo

import "fmt"

// Foo gets foo.
func Foo(x *string) (chan struct{}) {
	panic("hello, world")
}
`
	case ".js":
		return `import { foo } from 'bar'

export function baz(n) {
	return document.getElementById('x')
}
`
	case ".m":
		return `@interface X:NSObject {
	double x;
}

@property(nonatomic, readwrite) double foo;`
	default:
		return ""
	}
}

func mustMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
