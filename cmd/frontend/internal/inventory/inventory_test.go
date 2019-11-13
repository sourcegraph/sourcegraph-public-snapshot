package inventory

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/src-d/enry/v2"
)

func TestGetLang_language(t *testing.T) {
	tests := map[string]struct {
		file fi
		want string
	}{
		"empty file": {file: fi{"a", ""}, want: ""},
		"java":       {file: fi{"a.java", "a"}, want: "Java"},
		"go":         {file: fi{"a.go", "a"}, want: "Go"},

		// Ensure that .tsx and .jsx are considered as valid extensions for TypeScript and JavaScript,
		// respectively.
		"override tsx": {file: fi{"a.tsx", "xx"}, want: "TypeScript"},
		"override jsx": {file: fi{"b.jsx", "x"}, want: "JavaScript"},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			lang, err := getLang(context.Background(), test.file, nil)
			if err != nil {
				t.Fatal(err)
			}
			if lang.Name != test.want {
				t.Fatalf("got %q, want %q", lang.Name, test.want)
			}
		})
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
	tests := []struct {
		file os.FileInfo
		want string
	}{
		{file: fi{"a.java", "aaaaaaaaa"}, want: "Java"},
		{file: fi{"b.md", "# Hello"}, want: "Markdown"},

		// The .m extension is used by many languages, but this code is obviously Objective-C. This
		// test checks that this file is detected correctly as Objective-C.
		{
			file: fi{"c.m", "@interface X:NSObject { double x; } @property(nonatomic, readwrite) double foo;"},
			want: "Objective-C",
		},
	}
	for _, test := range tests {
		t.Run(test.file.Name(), func(t *testing.T) {
			rc := ioutil.NopCloser(bytes.NewReader([]byte(test.file.(fi).Contents)))
			lang, err := getLang(context.Background(), test.file, rc)
			if err != nil {
				t.Fatal(err)
			}
			if lang.Name != test.want {
				t.Errorf("got %q, want %q", lang.Name, test.want)
			}
		})
	}
}

type nopReadCloser struct {
	r *bytes.Reader
}

func (n *nopReadCloser) Read(b []byte) (int, error) {
	return n.r.Read(b)
}

func (n *nopReadCloser) Close() error {
	return nil
}

func BenchmarkGetLang(b *testing.B) {
	files, err := readFileTree("prom-repo-tree.txt")
	if err != nil {
		b.Fatal(err)
	}
	dataMap := make(map[string][]byte)
	for _, f := range files {
		dataMap[f.Name()] = []byte(f.(fi).Contents)
	}
	b.Logf("Calling Get on %d files.", len(files))
	rc := &nopReadCloser{
		r: bytes.NewReader(nil),
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, file := range files {
			data := dataMap[file.Name()]
			rc.r.Reset(data)
			_, err = getLang(context.Background(), file, rc)
			if err != nil {
				b.Fatal(err)
			}
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
