package langserver

import (
	"fmt"
	"go/build"
	"go/token"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

var loaderCases = map[string]struct {
	fs map[string]string
}{
	"standalone": {fs: map[string]string{"/src/p/f.go": `package p; func F() {}`}},
	"imports net/http": {
		fs: map[string]string{"/src/p/f.go": `package p; import "net/http"; var _ = http.Get`},
	},
	"build-tagged different package in dir": {
		fs: map[string]string{
			"/src/p/f.go": `package p`,
			"/src/p/main.go": `// +build ignore

package main`,
		},
	},
	"multiple packages in dir": {
		fs: map[string]string{
			"/src/p/f.go":    `package p`,
			"/src/p/main.go": `package main`,
		},
	},
}

func TestLoader(t *testing.T) {
	for label, tc := range loaderCases {
		t.Run(label, func(t *testing.T) {
			fset, bctx := setUpLoaderTest(tc.fs)
			if _, _, err := typecheck(fset, bctx, &build.Package{ImportPath: "p", Dir: "/src/p"}); err != nil {
				t.Error(err)
			}
		})
	}
}

// BenchmarkLoader measures the performance of loading and
// typechecking.
//
// Run it with:
//
//   go test ./xlang/golang -bench Loader -benchmem
func BenchmarkLoader(b *testing.B) {
	for label, tc := range loaderCases {
		b.Run(label, func(b *testing.B) {
			fset, bctx := setUpLoaderTest(tc.fs)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, _, err := typecheck(fset, bctx, &build.Package{ImportPath: "p", Dir: "/src/p"}); err != nil {
					b.Error(err)
				}
			}
		})
	}
}

func setUpLoaderTest(fs map[string]string) (*token.FileSet, *build.Context) {
	h := LangHandler{HandlerShared: new(HandlerShared)}
	if err := h.reset(&InitializeParams{
		InitializeParams:     lsp.InitializeParams{RootPath: "file:///src/p"},
		NoOSFileSystemAccess: true,
		BuildContext: &InitializeBuildContextParams{
			GOPATH: "/",
		},
	}); err != nil {
		panic(err)
	}
	for filename, contents := range fs {
		h.addOverlayFile("file://"+filename, []byte(contents))
	}
	bctx := h.OverlayBuildContext(nil, &build.Default, false)
	bctx.GOPATH = "/"
	return token.NewFileSet(), bctx
}

func TestBuildPackageForNamedFileInMultiPackageDir(t *testing.T) {
	tests := map[string]struct {
		bpkg *build.Package
		m    *build.MultiplePackageError
		want map[string]*build.Package // filename -> expected pkg
	}{
		"a and b": {
			bpkg: &build.Package{
				GoFiles:      []string{"a.go", "b.go"},
				TestGoFiles:  []string{"a_test.go", "b_test.go"},
				XTestGoFiles: []string{"xa_test.go", "xb_test.go"},
			},
			m: &build.MultiplePackageError{
				Packages: []string{"a", "a", "b", "b", "a_test", "b_test"},
				Files:    []string{"a.go", "a_test.go", "b.go", "b_test.go", "xa_test.go", "xb_test.go"},
			},
			want: map[string]*build.Package{
				"a.go":       &build.Package{Name: "a", GoFiles: []string{"a.go"}, TestGoFiles: []string{"a_test.go"}, XTestGoFiles: []string{"xa_test.go"}},
				"a_test.go":  &build.Package{Name: "a", GoFiles: []string{"a.go"}, TestGoFiles: []string{"a_test.go"}, XTestGoFiles: []string{"xa_test.go"}},
				"xa_test.go": &build.Package{Name: "a_test", GoFiles: []string{"a.go"}, TestGoFiles: []string{"a_test.go"}, XTestGoFiles: []string{"xa_test.go"}},
				"b.go":       &build.Package{Name: "b", GoFiles: []string{"b.go"}, TestGoFiles: []string{"b_test.go"}, XTestGoFiles: []string{"xb_test.go"}},
				"b_test.go":  &build.Package{Name: "b", GoFiles: []string{"b.go"}, TestGoFiles: []string{"b_test.go"}, XTestGoFiles: []string{"xb_test.go"}},
				"xb_test.go": &build.Package{Name: "b_test", GoFiles: []string{"b.go"}, TestGoFiles: []string{"b_test.go"}, XTestGoFiles: []string{"xb_test.go"}},
			},
		},
		"two main packages": {
			// TODO(sqs): If the package name is "main", and there are
			// multiple main packages that are separate programs (and,
			// e.g., expected to be run directly run `go run main1.go
			// main2.go`), then it will break because it will try to
			// compile them all together. There's no good way to handle
			// that case that I can think of, other than with heuristics.
			bpkg: &build.Package{
				GoFiles:     []string{"main1.go", "main2.go"},
				TestGoFiles: []string{"main_test.go"},
			},
			m: &build.MultiplePackageError{
				Packages: []string{"main", "main", "main"},
				Files:    []string{"main1.go", "main2.go", "main_test.go"},
			},
			want: map[string]*build.Package{
				"main1.go": &build.Package{Name: "main", GoFiles: []string{"main1.go", "main2.go"}, TestGoFiles: []string{"main_test.go"}},
				"main2.go": &build.Package{Name: "main", GoFiles: []string{"main1.go", "main2.go"}, TestGoFiles: []string{"main_test.go"}},
			},
		},
		"main with test": {
			bpkg: &build.Package{
				GoFiles:     []string{"a.go", "main.go"},
				TestGoFiles: []string{"main_test.go"},
			},
			m: &build.MultiplePackageError{
				Packages: []string{"a", "main", "main"},
				Files:    []string{"a.go", "main.go", "main_test.go"},
			},
			want: map[string]*build.Package{
				"a.go":         &build.Package{Name: "a", GoFiles: []string{"a.go"}},
				"main.go":      &build.Package{Name: "main", GoFiles: []string{"main.go"}, TestGoFiles: []string{"main_test.go"}},
				"main_test.go": &build.Package{Name: "main", GoFiles: []string{"main.go"}, TestGoFiles: []string{"main_test.go"}},
			},
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			for filename, want := range test.want {
				t.Run(filename, func(t *testing.T) {
					bpkg, err := buildPackageForNamedFileInMultiPackageDir(test.bpkg, test.m, filename)
					if err != nil {
						t.Fatalf("%s: %s: %s", label, filename, err)
					}
					if !reflect.DeepEqual(bpkg, want) {
						printPkg := func(p *build.Package) string {
							return fmt.Sprintf("build.Package{Name:%s GoFiles:%v TestGoFiles:%v XTestGoFiles:%v}", p.Name, p.GoFiles, p.TestGoFiles, p.XTestGoFiles)
						}
						t.Errorf("%s: %s:\n got %s\nwant %s", label, filename, printPkg(bpkg), printPkg(want))
					}
				})
			}
		})
	}
}
