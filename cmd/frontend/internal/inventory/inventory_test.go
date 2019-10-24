package inventory

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/src-d/enry/v2"
)

func TestDetect_noReadFile(t *testing.T) {
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
			lang, err := detect(context.Background(), test.file, nil)
			if err != nil {
				t.Fatal(err)
			}
			if lang != test.want {
				t.Fatalf("got %q, want %q", lang, test.want)
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

func newReadFile(files []os.FileInfo) func(_ context.Context, path string, minFileBytes int64) ([]byte, error) {
	m := make(map[string][]byte, len(files))
	for _, f := range files {
		m[f.Name()] = []byte(f.(fi).Contents)
	}
	return func(_ context.Context, path string, minFileBytes int64) ([]byte, error) {
		data, ok := m[path]
		if !ok {
			return nil, fmt.Errorf("no file: %s", path)
		}
		return data, nil
	}
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
			lang, err := detect(context.Background(), test.file, func(_ context.Context, path string, minFileBytes int64) ([]byte, error) {
				return []byte(test.file.(fi).Contents), nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if lang != test.want {
				t.Errorf("got %q, want %q", lang, test.want)
			}
		})
	}
}

func BenchmarkGet(b *testing.B) {
	files, err := readFileTree("prom-repo-tree.txt")
	if err != nil {
		b.Fatal(err)
	}
	readFile := newReadFile(files)
	b.Logf("Calling Get on %d files.", len(files))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, file := range files {
			_, err = detect(context.Background(), file, readFile)
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
