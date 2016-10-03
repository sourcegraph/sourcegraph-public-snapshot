package xlang_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/ctxvfs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/golang"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspx"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

func TestProxy(t *testing.T) {
	tests := map[string]struct {
		rootPath       string
		mode           string
		fs             map[string]string
		wantHover      map[string]string
		wantDefinition map[string]string
		wantReferences map[string][]string
		wantSymbols    map[string][]string
		depFS          map[string]map[string]string // dep clone URL -> map VFS
	}{
		"go basic": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: map[string]string{
				"a.go": "package p; func A() { A() }",
				"b.go": "package p; func B() { A() }",
			},
			wantHover: map[string]string{
				"a.go:1:9":  "package p",
				"a.go:1:17": "func A()",
				"a.go:1:23": "func A()",
				"b.go:1:17": "func B()",
				"b.go:1:23": "func A()",
			},
			wantDefinition: map[string]string{
				"a.go:1:17": "git://test/pkg?master#a.go:1:17",
				"a.go:1:23": "git://test/pkg?master#a.go:1:17",
				"b.go:1:17": "git://test/pkg?master#b.go:1:17",
				"b.go:1:23": "git://test/pkg?master#a.go:1:17",
			},
			wantReferences: map[string][]string{
				"a.go:1:17": []string{
					"git://test/pkg?master#a.go:1:17",
					"git://test/pkg?master#a.go:1:23",
					"git://test/pkg?master#b.go:1:23",
				},
				"a.go:1:23": []string{
					"git://test/pkg?master#a.go:1:17",
					"git://test/pkg?master#a.go:1:23",
					"git://test/pkg?master#b.go:1:23",
				},
				"b.go:1:17": []string{"git://test/pkg?master#b.go:1:17"},
				"b.go:1:23": []string{
					"git://test/pkg?master#a.go:1:17",
					"git://test/pkg?master#a.go:1:23",
					"git://test/pkg?master#b.go:1:23",
				},
			},
			wantSymbols: map[string][]string{
				"":  []string{"git://test/pkg?master#a.go:function:pkg.A", "git://test/pkg?master#b.go:function:pkg.B"},
				"A": []string{"git://test/pkg?master#a.go:function:pkg.A"},
				"B": []string{"git://test/pkg?master#b.go:function:pkg.B"},
			},
		},
		"go detailed": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: map[string]string{
				"a.go": "package p; type T struct { F string }",
			},
			wantHover: map[string]string{
			// "a.go:1:28": "(T).F string", // TODO(sqs): see golang/hover.go; this is the output we want
			},
			wantSymbols: map[string][]string{
				"":  []string{"git://test/pkg?master#a.go:class:pkg.T"},
				"T": []string{"git://test/pkg?master#a.go:class:pkg.T"},
				"F": []string{}, // we don't return fields for now
			},
		},
		"go xtest": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: map[string]string{
				"a.go":      "package p; var A int",
				"a_test.go": `package p_test; import "test/pkg"; var X = p.A`,
			},
			wantHover: map[string]string{
				"a.go:1:16":      "A int",
				"a_test.go:1:40": "X int",
				"a_test.go:1:46": "A int",
			},
		},
		"go subdirectory in repo": {
			rootPath: "git://test/pkg?master#d",
			mode:     "go",
			fs: map[string]string{
				"a.go":    "package d; func A() { A() }",
				"d2/b.go": `package d2; import "test/pkg/d"; func B() { d.A(); B() }`,
			},
			wantHover: map[string]string{
				"a.go:1:17":    "func A()",
				"a.go:1:23":    "func A()",
				"d2/b.go:1:39": "func B()",
				"d2/b.go:1:47": "func A()",
				"d2/b.go:1:52": "func B()",
			},
			wantDefinition: map[string]string{
				"a.go:1:17":    "git://test/pkg?master#d/a.go:1:17",
				"a.go:1:23":    "git://test/pkg?master#d/a.go:1:17",
				"d2/b.go:1:39": "git://test/pkg?master#d/d2/b.go:1:39",
				"d2/b.go:1:47": "git://test/pkg?master#d/a.go:1:17",
				"d2/b.go:1:52": "git://test/pkg?master#d/d2/b.go:1:39",
			},
			wantSymbols: map[string][]string{
				"": []string{"git://test/pkg?master#d/a.go:function:d.A", "git://test/pkg?master#d/d2/b.go:function:d2.B"},
			},
		},
		"go multiple packages in dir": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: map[string]string{
				"a.go": "package p; func A() { A() }",
				"main.go": `// +build ignore

package main; import "test/pkg"; func B() { p.A(); B() }`,
			},
			wantHover: map[string]string{
				"a.go:1:17": "func A()",
				"a.go:1:23": "func A()",
				// Not parsing build-tag-ignored files:
				//
				// "main.go:3:39": "func B()", // func B()
				// "main.go:3:47": "func A()", // p.A()
				// "main.go:3:52": "func B()", // B()
			},
			wantDefinition: map[string]string{
				"a.go:1:17": "git://test/pkg?master#a.go:1:17",
				"a.go:1:23": "git://test/pkg?master#a.go:1:17",
				// Not parsing build-tag-ignored files:
				//
				// "main.go:3:39": "git://test/pkg?master#main.go:3:39", // B() -> func B()
				// "main.go:3:47": "git://test/pkg?master#a.go:1:17",    // p.A() -> a.go func A()
				// "main.go:3:52": "git://test/pkg?master#main.go:3:39", // B() -> func B()
			},
			wantSymbols: map[string][]string{
				"": []string{"git://test/pkg?master#a.go:function:pkg.A"},
			},
		},
		"goroot": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: map[string]string{
				"a.go": `package p; import "fmt"; var _ = fmt.Println; var x int`,
			},
			wantHover: map[string]string{
				"a.go:1:40": "func Println(a ...interface{}) (n int, err error)",
				"a.go:1:53": "type int int",
			},
			wantDefinition: map[string]string{
				"a.go:1:40": "git://github.com/golang/go?" + runtime.Version() + "#src/fmt/print.go:1:19",
				// "a.go:1:53": "git://github.com/golang/go?" + runtime.Version() + "#src/builtin/builtin.go:TODO:TODO", // TODO(sqs): support builtins
			},
			depFS: map[string]map[string]string{
				"https://github.com/golang/go?go1.7.1": {
					"src/fmt/print.go":       "package fmt; func Println(a ...interface{}) (n int, err error) { return }",
					"src/builtin/builtin.go": "package builtin; type int int",
				},
			},
			wantSymbols: map[string][]string{
				"": []string{
					"git://test/pkg?master#a.go:variable:pkg._",
					"git://test/pkg?master#a.go:variable:pkg.x",
				},
			},
		},
		"gopath": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: map[string]string{
				"a/a.go": `package a; func A() {}`,
				"b/b.go": `package b; import "test/pkg/a"; var _ = a.A`,
			},
			wantHover: map[string]string{
				"a/a.go:1:17": "func A()",
				// "b/b.go:1:20": "package", // TODO(sqs): make import paths hoverable
				"b/b.go:1:43": "func A()",
			},
			wantDefinition: map[string]string{
				"a/a.go:1:17": "git://test/pkg?master#a/a.go:1:17",
				// "b/b.go:1:20": "git://test/pkg?master#a", // TODO(sqs): make import paths hoverable
				"b/b.go:1:43": "git://test/pkg?master#a/a.go:1:17",
			},
			wantReferences: map[string][]string{
				"a/a.go:1:17": []string{
					"git://test/pkg?master#a/a.go:1:17",
					"git://test/pkg?master#b/b.go:1:43",
				},
				"b/b.go:1:43": []string{ // calling "references" on call site should return same result as on decl
					"git://test/pkg?master#a/a.go:1:17",
					"git://test/pkg?master#b/b.go:1:43",
				},
			},
			wantSymbols: map[string][]string{
				"": []string{"git://test/pkg?master#a/a.go:function:a.A", "git://test/pkg?master#b/b.go:variable:b._"},
			},
		},
		"go vendored dep": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: map[string]string{
				"a.go": `package a; import "github.com/v/vendored"; var _ = vendored.V`,
				"vendor/github.com/v/vendored/v.go": "package vendored; func V() {}",
			},
			wantHover: map[string]string{
				"a.go:1:61": "func V()",
			},
			wantDefinition: map[string]string{
				"a.go:1:61": "git://test/pkg?master#vendor/github.com/v/vendored/v.go:1:24",
			},
			wantReferences: map[string][]string{
				"vendor/github.com/v/vendored/v.go:1:24": []string{
					"git://test/pkg?master#vendor/github.com/v/vendored/v.go:1:24",
					"git://test/pkg?master#a.go:1:61",
				},
			},
			wantSymbols: map[string][]string{
				"": []string{"git://test/pkg?master#a.go:variable:pkg._", "git://test/pkg?master#vendor/github.com/v/vendored/v.go:function:vendored.V"},
			},
		},
		"go vendor symbols with same name": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: map[string]string{
				"z.go": `package pkg; func x() bool { return true }`,
				"vendor/github.com/a/pkg2/x.go": `package pkg2; func x() bool { return true }`,
				"vendor/github.com/x/pkg3/x.go": `package pkg3; func x() bool { return true }`,
			},
			wantSymbols: map[string][]string{
				"": []string{
					"git://test/pkg?master#z.go:function:pkg.x",
					"git://test/pkg?master#vendor/github.com/a/pkg2/x.go:function:pkg2.x",
					"git://test/pkg?master#vendor/github.com/x/pkg3/x.go:function:pkg3.x",
				},
				"x": []string{
					"git://test/pkg?master#z.go:function:pkg.x",
					"git://test/pkg?master#vendor/github.com/a/pkg2/x.go:function:pkg2.x",
					"git://test/pkg?master#vendor/github.com/x/pkg3/x.go:function:pkg3.x",
				},
				"pkg2.x": []string{
					"git://test/pkg?master#vendor/github.com/a/pkg2/x.go:function:pkg2.x",
					"git://test/pkg?master#z.go:function:pkg.x",
					"git://test/pkg?master#vendor/github.com/x/pkg3/x.go:function:pkg3.x",
				},
				"pkg3.x": []string{
					"git://test/pkg?master#vendor/github.com/x/pkg3/x.go:function:pkg3.x",
					"git://test/pkg?master#z.go:function:pkg.x",
					"git://test/pkg?master#vendor/github.com/a/pkg2/x.go:function:pkg2.x",
				},
			},
		},
		"go external dep": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: map[string]string{
				"a.go": `package a; import "github.com/d/dep"; var _ = dep.D; var _ = dep.D`,
			},
			wantHover: map[string]string{
				"a.go:1:51": "func D()",
			},
			wantDefinition: map[string]string{
				"a.go:1:51": "git://github.com/d/dep?HEAD#d.go:1:19",
			},
			wantReferences: map[string][]string{
				"a.go:1:51": []string{
					"git://test/pkg?master#a.go:1:51",
					"git://test/pkg?master#a.go:1:66",
					// Do not include "refs" from the dependency
					// package itself; only return results in the
					// workspace.
				},
			},
			depFS: map[string]map[string]string{
				"https://github.com/d/dep?HEAD": {
					"d.go": "package dep; func D() {}; var _ = D",
				},
			},
		},
		"external dep with vendor": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: map[string]string{
				"a.go": `package p; import "github.com/d/dep"; var _ = dep.D().F`,
			},
			wantDefinition: map[string]string{
				"a.go:1:55": "git://github.com/d/dep?HEAD#vendor/vendp/vp.go:1:32",
			},
			depFS: map[string]map[string]string{
				"https://github.com/d/dep?HEAD": map[string]string{
					"d.go":               `package dep; import "vendp"; func D() (v vendp.V) { return }`,
					"vendor/vendp/vp.go": "package vendp; type V struct { F int }",
				},
			},
		},
		"go external dep at subtree": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: map[string]string{
				"a.go": `package a; import "github.com/d/dep/subp"; var _ = subp.D`,
			},
			wantHover: map[string]string{
				"a.go:1:57": "func D()",
			},
			wantDefinition: map[string]string{
				"a.go:1:57": "git://github.com/d/dep?HEAD#subp/d.go:1:20",
			},
			depFS: map[string]map[string]string{
				"https://github.com/d/dep?HEAD": {
					"subp/d.go": "package subp; func D() {}",
				},
			},
		},
		"go nested external dep": { // a depends on dep1, dep1 depends on dep2
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: map[string]string{
				"a.go": `package a; import "github.com/d/dep1"; var _ = dep1.D1().D2`,
			},
			wantHover: map[string]string{
				"a.go:1:53": "func D1() D2",
				"a.go:1:59": "D2 int",
			},
			wantDefinition: map[string]string{
				"a.go:1:53": "git://github.com/d/dep1?HEAD#d1.go:1:48", // func D1
				"a.go:1:58": "git://github.com/d/dep2?HEAD#d2.go:1:32", // field D2
			},
			depFS: map[string]map[string]string{
				"https://github.com/d/dep1?HEAD": {
					"d1.go": `package dep1; import "github.com/d/dep2"; func D1() dep2.D2 { return dep2.D2{} }`,
				},
				"https://github.com/d/dep2?HEAD": {
					"d2.go": "package dep2; type D2 struct { D2 int }",
				},
			},
		},
		"go external dep at vanity import path": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: map[string]string{
				"a.go": `package a; import "golang.org/x/text"; var _ = text.F`,
			},
			wantHover: map[string]string{
				"a.go:1:53": "func F()",
			},
			wantDefinition: map[string]string{
				"a.go:1:53": "git://github.com/golang/text?HEAD#dummy.go:1:20",
			},
			depFS: map[string]map[string]string{
				// We override the Git cloning of this repo to use
				// in-memory dummy data, but we still need to hit the
				// network to resolve the Go custom import path
				// (because that's not mocked yet).
				"https://github.com/golang/text?HEAD": {
					"dummy.go": "package text; func F() {}",
				},
			},
		},

		// This covers repos like github.com/kubernetes/kubernetes,
		// which have doc.go files in subpackages with canonical
		// import path comments of "//
		// k8s.io/kubernetes/SUBPACKAGE...". If we don't set up the
		// workspace at /src/k8s.io/kubernetes, then cross-package
		// definitions will fail, and we will erroneously fetch a
		// separate (HEAD) copy of the entire kubernetes repo at the
		// k8s.io/kubernetes/... root.
		"go packages with canonical import path different from its repo": {
			rootPath: "git://test/foo?master",
			mode:     "go",
			fs: map[string]string{
				"a/a.go": `package a // import "other/foo/a"

import "other/foo/b"

var A = b.B`,
				"b/b.go": `package b // import "other/foo/b"

var (
	B = 123
	bb = B
)`,
			},
			wantDefinition: map[string]string{
				"a/a.go:5:5":  "git://test/foo?master#a/a.go:5:5", // "var A"
				"a/a.go:5:11": "git://test/foo?master#b/b.go:4:2", // "b.B"
				"b/b.go:4:2":  "git://test/foo?master#b/b.go:4:2", // "B = 123"
				"b/b.go:5:7":  "git://test/foo?master#b/b.go:4:2", // "bb = B"
			},
		},

		"go symbols": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: map[string]string{
				"abc.go": `package a; type XYZ struct {}; func (x XYZ) ABC() {}`,
				"bcd.go": `package a; type YZA struct {}; func (y YZA) BCD() {}`,
				"xyz.go": `package a; func yza() {}`,
			},
			wantSymbols: map[string][]string{
				"":    []string{"git://test/pkg?master#abc.go:method:XYZ.ABC", "git://test/pkg?master#bcd.go:method:YZA.BCD", "git://test/pkg?master#abc.go:class:pkg.XYZ", "git://test/pkg?master#bcd.go:class:pkg.YZA", "git://test/pkg?master#xyz.go:function:pkg.yza"},
				"xyz": []string{"git://test/pkg?master#abc.go:class:pkg.XYZ", "git://test/pkg?master#abc.go:method:XYZ.ABC", "git://test/pkg?master#xyz.go:function:pkg.yza"},
				"yza": []string{"git://test/pkg?master#bcd.go:class:pkg.YZA", "git://test/pkg?master#xyz.go:function:pkg.yza", "git://test/pkg?master#bcd.go:method:YZA.BCD"},
				"abc": []string{"git://test/pkg?master#abc.go:method:XYZ.ABC", "git://test/pkg?master#abc.go:class:pkg.XYZ"},
				"bcd": []string{"git://test/pkg?master#bcd.go:method:YZA.BCD", "git://test/pkg?master#bcd.go:class:pkg.YZA"},
			},
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			// Mock repo and dep fetching to use test fixtures.
			{
				orig := xlang.NewRemoteRepoVFS
				xlang.NewRemoteRepoVFS = func(cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
					return mapFS(test.fs), nil
				}
				defer func() {
					xlang.NewRemoteRepoVFS = orig
				}()
			}
			{
				orig := golang.NewDepRepoVFS
				golang.NewDepRepoVFS = func(cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
					id := cloneURL.String() + "?" + rev
					if fs, ok := test.depFS[id]; ok {
						return mapFS(fs), nil
					}
					return nil, fmt.Errorf("no file system found for dep at %s rev %q", cloneURL, rev)
				}
				defer func() {
					golang.NewDepRepoVFS = orig
				}()
			}

			ctx := context.Background()
			proxy := xlang.NewProxy()
			if test.rootPath == "" {
				t.Fatal("no rootPath set in test fixture")
			}

			addr, done := startProxy(t, proxy)
			defer done()
			c := dialProxy(t, addr, nil)

			root, err := uri.Parse(test.rootPath)
			if err != nil {
				t.Fatal(err)
			}

			// Prepare the connection.
			if err := c.Call(ctx, "initialize", xlang.ClientProxyInitializeParams{
				InitializeParams: lsp.InitializeParams{RootPath: test.rootPath},
				Mode:             test.mode,
			}, nil); err != nil {
				t.Fatal("initialize:", err)
			}

			for pos, want := range test.wantHover {
				t.Run(fmt.Sprintf("hover-%s", pos), func(t *testing.T) {
					hoverTest(t, ctx, c, root, pos, want)
				})
			}

			for pos, want := range test.wantDefinition {
				t.Run(fmt.Sprintf("definition-%s", pos), func(t *testing.T) {
					definitionTest(t, ctx, c, root, pos, want)
				})
			}

			for pos, want := range test.wantReferences {
				t.Run(fmt.Sprintf("references-%s", pos), func(t *testing.T) {
					referencesTest(t, ctx, c, root, pos, want)
				})
			}

			for query, want := range test.wantSymbols {
				t.Run(fmt.Sprintf("symbols(q=%q)", query), func(t *testing.T) {
					symbolsTest(t, ctx, c, root, query, want)
				})
			}
		})
	}
}

func startProxy(t testing.TB, proxy *xlang.Proxy) (addr string, done func()) {
	bindAddr := ":0"
	if os.Getenv("CI") != "" {
		// CircleCI has issues with IPv6 (e.g., "dial tcp [::]:39984:
		// connect: network is unreachable").
		bindAddr = "127.0.0.1:0"
	}
	l, err := net.Listen("tcp", bindAddr)
	if err != nil {
		t.Fatal("Listen:", err)
	}
	go func() {
		if err := proxy.Serve(context.Background(), l); err != nil {
			t.Fatal("proxy.Serve:", err)
		}
	}()
	return l.Addr().String(), func() {
		if err := proxy.Close(context.Background()); err != nil && err.Error() != "jsonrpc2: connection is closed" {
			t.Fatal("proxy.Close:", err)
		}
	}
}

func dialProxy(t testing.TB, addr string, recvDiags chan<- lsp.PublishDiagnosticsParams) *jsonrpc2.Conn {
	h := &xlang.ClientHandler{
		RecvDiagnostics: func(uri string, diags []lsp.Diagnostic) {
			if recvDiags == nil {
				var buf bytes.Buffer
				for _, d := range diags {
					fmt.Fprintf(&buf, "\t:%d:%d: %s\n", d.Range.Start.Line+1, d.Range.Start.Character+1, d.Message)
				}
				t.Logf("diagnostics: %s\n%s", uri, buf.String())
			} else {
				recvDiags <- lsp.PublishDiagnosticsParams{URI: uri, Diagnostics: diags}
			}
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	c, err := xlang.DialProxy(ctx, addr, h)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func hoverTest(t *testing.T, ctx context.Context, c *jsonrpc2.Conn, root *uri.URI, pos, want string) {
	file, line, char, err := parsePos(pos)
	if err != nil {
		t.Fatal(err)
	}
	hover, err := callHover(ctx, c, root.WithFilePath(file).String(), line, char)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasSuffix(want, "...") {
		// Allow specifying expected hover strings with "..." at the
		// end for ease of test creation.
		if len(hover) >= len(want)+3 {
			hover = hover[:len(want)-3] + "..."
		}
	}
	if !strings.Contains(hover, want) {
		t.Fatalf("got %q, want %q", hover, want)
	}
}

func definitionTest(t testing.TB, ctx context.Context, c *jsonrpc2.Conn, root *uri.URI, pos, want string) {
	file, line, char, err := parsePos(pos)
	if err != nil {
		t.Fatal(err)
	}
	definition, err := callDefinition(ctx, c, root.WithFilePath(file).String(), line, char)
	if err != nil {
		t.Fatal(err)
	}
	definition = strings.TrimPrefix(definition, "file:///")
	if definition != want {
		t.Errorf("got %q, want %q", definition, want)
	}
}

func referencesTest(t testing.TB, ctx context.Context, c *jsonrpc2.Conn, root *uri.URI, pos string, want []string) {
	file, line, char, err := parsePos(pos)
	if err != nil {
		t.Fatal(err)
	}
	references, err := callReferences(ctx, c, root.WithFilePath(file).String(), line, char)
	if err != nil {
		t.Fatal(err)
	}
	for i := range references {
		references[i] = strings.TrimPrefix(references[i], "file:///")
	}
	sort.Strings(references)
	sort.Strings(want)
	if !reflect.DeepEqual(references, want) {
		t.Errorf("got %q, want %q", references, want)
	}
}

func symbolsTest(t testing.TB, ctx context.Context, c *jsonrpc2.Conn, root *uri.URI, query string, want []string) {
	symbols, err := callSymbols(ctx, c, query)
	if err != nil {
		t.Fatal(err)
	}
	for i := range symbols {
		symbols[i] = strings.TrimPrefix(symbols[i], "file:///")
	}
	if !reflect.DeepEqual(symbols, want) {
		t.Errorf("got %#v, want %q", symbols, want)
	}
}

func parsePos(s string) (file string, line, char int, err error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		err = fmt.Errorf("invalid pos %q (%d parts)", s, len(parts))
		return
	}
	file = parts[0]
	line, err = strconv.Atoi(parts[1])
	if err != nil {
		err = fmt.Errorf("invalid line in %q: %s", s, err)
		return
	}
	char, err = strconv.Atoi(parts[2])
	if err != nil {
		err = fmt.Errorf("invalid char in %q: %s", s, err)
		return
	}
	return file, line - 1, char - 1, nil // LSP is 0-indexed
}

func callHover(ctx context.Context, c *jsonrpc2.Conn, uri string, line, char int) (string, error) {
	var res struct {
		Contents markedStrings `json:"contents"`
		lsp.Hover
	}
	err := c.Call(ctx, "textDocument/hover", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{Line: line, Character: char},
	}, &res)
	if err != nil {
		return "", err
	}
	var str string
	for i, ms := range res.Contents {
		if i != 0 {
			str += " "
		}
		str += ms.Value
	}
	return str, nil
}

func callDefinition(ctx context.Context, c *jsonrpc2.Conn, uri string, line, char int) (string, error) {
	var res locations
	err := c.Call(ctx, "textDocument/definition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{Line: line, Character: char},
	}, &res)
	if err != nil {
		return "", err
	}
	var str string
	for i, loc := range res {
		if loc.URI == "" {
			continue
		}
		if i != 0 {
			str += ", "
		}
		str += fmt.Sprintf("%s:%d:%d", loc.URI, loc.Range.Start.Line+1, loc.Range.Start.Character+1)
	}
	return str, nil
}

func callReferences(ctx context.Context, c *jsonrpc2.Conn, uri string, line, char int) ([]string, error) {
	var res locations
	err := c.Call(ctx, "textDocument/references", lsp.ReferenceParams{
		Context: lsp.ReferenceContext{IncludeDeclaration: true},
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: uri},
			Position:     lsp.Position{Line: line, Character: char},
		},
	}, &res)
	if err != nil {
		return nil, err
	}
	str := make([]string, len(res))
	for i, loc := range res {
		str[i] = fmt.Sprintf("%s:%d:%d", loc.URI, loc.Range.Start.Line+1, loc.Range.Start.Character+1)
	}
	return str, nil
}

var symbolKindName = map[lsp.SymbolKind]string{
	lsp.SKFile:        "file",
	lsp.SKModule:      "module",
	lsp.SKNamespace:   "namespace",
	lsp.SKPackage:     "package",
	lsp.SKClass:       "class",
	lsp.SKMethod:      "method",
	lsp.SKProperty:    "property",
	lsp.SKField:       "field",
	lsp.SKConstructor: "constructor",
	lsp.SKEnum:        "enum",
	lsp.SKInterface:   "interface",
	lsp.SKFunction:    "function",
	lsp.SKVariable:    "variable",
	lsp.SKConstant:    "constant",
	lsp.SKString:      "string",
	lsp.SKNumber:      "number",
	lsp.SKBoolean:     "boolean",
	lsp.SKArray:       "array",
}

func callSymbols(ctx context.Context, c *jsonrpc2.Conn, query string) ([]string, error) {
	var symbols []lsp.SymbolInformation
	err := c.Call(ctx, "workspace/symbol", lsp.WorkspaceSymbolParams{Query: query}, &symbols)
	if err != nil {
		return nil, err
	}
	syms := make([]string, len(symbols))
	for i, s := range symbols {
		syms[i] = fmt.Sprintf("%s:%s:%s.%s", s.Location.URI, symbolKindName[s.Kind], s.ContainerName, s.Name)
	}
	return syms, nil
}

type markedStrings []lsp.MarkedString

func (v *markedStrings) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("invalid empty JSON")
	}
	if data[0] == '[' {
		var ms []markedString
		if err := json.Unmarshal(data, &ms); err != nil {
			return err
		}
		for _, ms := range ms {
			*v = append(*v, lsp.MarkedString(ms))
		}
		return nil
	}
	*v = []lsp.MarkedString{{}}
	return json.Unmarshal(data, &(*v)[0])
}

type markedString lsp.MarkedString

func (v *markedString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("invalid empty JSON")
	}
	if data[0] == '{' {
		return json.Unmarshal(data, (*lsp.MarkedString)(v))
	}

	// String
	*v = markedString{}
	return json.Unmarshal(data, &v.Value)
}

type locations []lsp.Location

func (v *locations) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("invalid empty JSON")
	}
	if data[0] == '[' {
		return json.Unmarshal(data, (*[]lsp.Location)(v))
	}
	*v = []lsp.Location{{}}
	return json.Unmarshal(data, &(*v)[0])
}

// TestProxy_connections tests that connections are created, reused
// and resumed as appropriate.
func TestProxy_connections(t *testing.T) {
	ctx := context.Background()

	orig := xlang.NewRemoteRepoVFS
	xlang.NewRemoteRepoVFS = func(cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
		return ctxvfs.Map(map[string][]byte{"f": []byte("x")}), nil
	}
	defer func() {
		xlang.NewRemoteRepoVFS = orig
	}()

	// Store data sent/received for checking.
	var (
		mu     sync.Mutex
		reqs   []testRequest // store received reqs
		addReq = func(req *jsonrpc2.Request) {
			mu.Lock()
			defer mu.Unlock()
			reqs = append(reqs, testRequest{req.Method, req.Params})
		}
		waitForReqs = func() error {
			for t0 := time.Now(); time.Since(t0) < 250*time.Millisecond; time.Sleep(10 * time.Millisecond) {
				mu.Lock()
				if reqs != nil {
					mu.Unlock()
					return nil
				}
				mu.Unlock()
			}
			return errors.New("timed out waiting for test server to receive req")
		}
		getAndClearReqs = func() []testRequest {
			mu.Lock()
			defer mu.Unlock()
			v := reqs
			reqs = nil
			return v
		}
		wantReqs = func(want []testRequest) error {
			if os.Getenv("CI") != "" {
				time.Sleep(3 * time.Second)
			}
			if err := waitForReqs(); err != nil {
				return err
			}
			got := getAndClearReqs()
			join := func(reqs []testRequest) (s string) {
				for i, r := range reqs {
					if i != 0 {
						s += "\n"
					}
					s += r.String()
				}
				return
			}
			sort.Sort(testRequests(got))
			sort.Sort(testRequests(want))
			if !testRequestsEqual(got, want) {
				return fmt.Errorf("got reqs != want reqs\n\nGOT REQS:\n%s\n\nWANT REQS:\n%s", join(got), join(want))
			}
			return nil
		}
	)

	// Start test build/lang server S1.
	calledConnectToTestServer := 0 // track the times we need to open a new server connection
	xlang.ServersByMode["test"] = func() (io.ReadWriteCloser, error) {
		mu.Lock()
		calledConnectToTestServer++
		mu.Unlock()
		a, b := xlang.InMemoryPeerConns()
		jsonrpc2.NewConn(context.Background(), a, jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
			addReq(req)
			return nil, nil
		}))
		return b, nil
	}
	defer func() {
		delete(xlang.ServersByMode, "test")
	}()

	proxy := xlang.NewProxy()
	addr, done := startProxy(t, proxy)
	defer done()

	// Start the test client C1.
	c1 := dialProxy(t, addr, nil)

	// C1 connects to the proxy.
	initParams := xlang.ClientProxyInitializeParams{
		InitializeParams: lsp.InitializeParams{RootPath: "test://test?v"},
		Mode:             "test",
	}
	if err := c1.Call(ctx, "initialize", initParams, nil); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond) // we're testing for a negative, so this is not as flaky as it seems; if a request is received later, it'll cause a test failure the next time we call wantReqs
	if got := getAndClearReqs(); len(got) != 0 {
		t.Errorf(`after C1 initialize, got reqs %s, want none

Nothing should've been received by S1 yet, since the "initialize" request is proxied and not delivered to S1 until it is needed for responding to an actual request. This may change in the future if we want to pre-warm it upon receiving the initialize, though.`, got)
	}

	// Now C1 sends an actual request. The proxy should open a
	// connection to S1, initialize it, and send the request.
	if err := c1.Call(ctx, "textDocument/definition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://test?v#myfile"},
		Position:     lsp.Position{Line: 1, Character: 2},
	}, nil); err != nil {
		t.Fatal(err)
	}
	want := []testRequest{
		{"initialize", lspx.InitializeParams{
			InitializeParams: lsp.InitializeParams{RootPath: "file:///"},
			OriginalRootPath: "test://test?v"}},
		{"textDocument/definition", lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///myfile"},
			Position:     lsp.Position{Line: 1, Character: 2}}},
	}
	if err := wantReqs(want); err != nil {
		t.Fatal("after C1 textDocument/definition request:", err)
	}
	if want := 1; calledConnectToTestServer != want {
		t.Errorf("got %d server connections, want %d (the server should have been connected to)", calledConnectToTestServer, want)
	}

	// C1 sends another request. The server is already initialized, so
	// just the single request needs to get sent.
	if err := c1.Call(ctx, "textDocument/hover", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://test?v#myfile2"},
		Position:     lsp.Position{Line: 3, Character: 4},
	}, nil); err != nil {
		t.Fatal(err)
	}
	want = []testRequest{
		{"textDocument/hover", lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///myfile2"},
			Position:     lsp.Position{Line: 3, Character: 4}}},
	}
	if err := wantReqs(want); err != nil {
		t.Fatal("after C1 textDocument/hover request:", err)
	}

	// Kill the server to simulate either an idle shutdown by the
	// proxy, or an unexpected failure on the server.
	if err := proxy.ShutDownIdleServers(context.Background(), 0 /* 0 means kill all */); err != nil {
		t.Fatal(err)
	}
	if want := 1; calledConnectToTestServer != want {
		t.Errorf("got %d server connections, want %d (after shutting down the server, did not expect proxy to reconnect until the next client request arrived that should be routed to the server)", calledConnectToTestServer, want)
	}

	// C1 does not know the server was killed. When it sends the next
	// request, the proxy should transparently spin up a new server
	// and reinitialize it appropriately.
	if err := c1.Call(ctx, "textDocument/definition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://test?v#myfile3"},
		Position:     lsp.Position{Line: 5, Character: 6},
	}, nil); err != nil {
		t.Fatal(err)
	}
	want = []testRequest{
		{"shutdown", nil},
		{"exit", nil},
		{"initialize", lspx.InitializeParams{
			InitializeParams: lsp.InitializeParams{RootPath: "file:///"},
			OriginalRootPath: "test://test?v"}},
		{"textDocument/definition", lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///myfile3"},
			Position:     lsp.Position{Line: 5, Character: 6}}},
	}
	if err := wantReqs(want); err != nil {
		t.Fatal("after C1's post-server-shutdown textDocument/definition request:", err)
	}
	if want := 2; calledConnectToTestServer != want {
		t.Errorf("got %d server connections, want %d (the server should have been reconnected to and reinitialized)", calledConnectToTestServer, want)
	}
}

// testRequest is a simplified version of jsonrpc2.Request for easier
// test expectation definition and checking of the fields that matter.
type testRequest struct {
	Method string
	Params interface{}
}

func (r testRequest) String() string {
	b, err := json.Marshal(r.Params)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s(%s)", r.Method, b)
}

func testRequestEqual(a, b testRequest) bool {
	if a.Method != b.Method {
		return false
	}

	// We want to see if a and b have identical canonical JSON
	// representations. They are NOT identical Go structures, since
	// one comes from the wire (as raw JSON) and one is an interface{}
	// of a concrete struct/slice type provided as a test expectation.
	ajson, err := json.Marshal(a.Params)
	if err != nil {
		panic(err)
	}
	bjson, err := json.Marshal(b.Params)
	if err != nil {
		panic(err)
	}
	var a2, b2 interface{}
	if err := json.Unmarshal(ajson, &a2); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(bjson, &b2); err != nil {
		panic(err)
	}
	return reflect.DeepEqual(a2, b2)
}

func testRequestsEqual(as, bs []testRequest) bool {
	if len(as) != len(bs) {
		return false
	}
	for i, a := range as {
		if !testRequestEqual(a, bs[i]) {
			return false
		}
	}
	return true
}

// TestProxy_propagation tests that diagnostics and log messages are
// propagated from the build/lang server through the proxy to the
// client.
func TestProxy_propagation(t *testing.T) {
	ctx := context.Background()

	orig := xlang.NewRemoteRepoVFS
	xlang.NewRemoteRepoVFS = func(cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
		return ctxvfs.Map(map[string][]byte{"f": []byte("x")}), nil
	}
	defer func() {
		xlang.NewRemoteRepoVFS = orig
	}()

	proxy := xlang.NewProxy()
	addr, done := startProxy(t, proxy)
	defer done()

	// Start test build/lang server that sends diagnostics about any
	// file that we call textDocument/definition on.
	xlang.ServersByMode["test"] = func() (io.ReadWriteCloser, error) {
		a, b := xlang.InMemoryPeerConns()
		jsonrpc2.NewConn(context.Background(), a, jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
			if req.Method == "textDocument/definition" {
				var params lsp.TextDocumentPositionParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				if err := conn.Notify(ctx, "textDocument/publishDiagnostics", lsp.PublishDiagnosticsParams{
					URI: params.TextDocument.URI,
					Diagnostics: []lsp.Diagnostic{
						{
							Range:   lsp.Range{Start: lsp.Position{Line: 1, Character: 1}, End: lsp.Position{Line: 1, Character: 1}},
							Message: "m",
						},
					},
				}); err != nil {
					return nil, err
				}
				return []lsp.Location{}, nil
			}
			return nil, nil
		}))
		return b, nil
	}
	defer func() {
		delete(xlang.ServersByMode, "test")
	}()

	recvDiags := make(chan lsp.PublishDiagnosticsParams)
	c := dialProxy(t, addr, recvDiags)

	// Connect to the proxy.
	initParams := xlang.ClientProxyInitializeParams{
		InitializeParams: lsp.InitializeParams{RootPath: "test://test?v"},
		Mode:             "test",
	}
	if err := c.Call(ctx, "initialize", initParams, nil); err != nil {
		t.Fatal(err)
	}

	// Call something that triggers the server to return diagnostics.
	if err := c.Call(ctx, "textDocument/definition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://test?v#myfile"},
		Position:     lsp.Position{Line: 1, Character: 2},
	}, nil); err != nil {
		t.Fatal(err)
	}

	// Check that we got the diagnostics.
	select {
	case diags := <-recvDiags:
		want := lsp.PublishDiagnosticsParams{
			URI: "test://test?v#myfile",
			Diagnostics: []lsp.Diagnostic{
				{
					Range:   lsp.Range{Start: lsp.Position{Line: 1, Character: 1}, End: lsp.Position{Line: 1, Character: 1}},
					Message: "m",
				},
			},
		}
		if !reflect.DeepEqual(diags, want) {
			t.Errorf("got diags\n%+v\n\nwant diags\n%+v", diags, want)
		}

	case <-time.After(time.Second):
		t.Fatal("want diagnostics, got nothing before timeout")
	}
}

// TestClientProxy_enforceAllURIsUnderneathRootPath tests that the
// client proxy forbids the use of any URIs in requests that are not
// underneath the initialize's rootPath. This is important for
// security as otherwise there is a risk that code could be fetched
// from other private repositories. This check is not the only
// safeguard (and without this safeguard, it would still forbid access
// to other repositories); this check is intended to increase the
// number of mistakes we need to make to introduce a security
// vulnerability.
func TestClientProxy_enforceAllURIsUnderneathRootPath(t *testing.T) {
	ctx := context.Background()

	proxy := xlang.NewProxy()
	addr, done := startProxy(t, proxy)
	defer done()

	c := dialProxy(t, addr, nil)

	// Connect to the proxy.
	initParams := xlang.ClientProxyInitializeParams{
		InitializeParams: lsp.InitializeParams{RootPath: "test://test?v"},
		Mode:             "test",
	}
	if err := c.Call(ctx, "initialize", initParams, nil); err != nil {
		t.Fatal(err)
	}

	// Send a request with a URI referring to a different repo from
	// the one in initialize's rootPath.
	if err := c.Call(ctx, "textDocument/definition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://different-repo#myfile"},
		Position:     lsp.Position{Line: 1, Character: 2},
	}, nil); err == nil || !strings.Contains(err.Error(), "must be underneath root path") {
		t.Fatalf("got error %v, want it to contain 'must be underneath root path'", err)
	}
}

type testRequests []testRequest

func (v testRequests) Len() int      { return len(v) }
func (v testRequests) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v testRequests) Less(i, j int) bool {
	ii, err := json.Marshal(v[i])
	if err != nil {
		panic(err)
	}
	jj, err := json.Marshal(v[j])
	if err != nil {
		panic(err)
	}
	return string(ii) < string(jj)
}

// mapFS lets us easily instantiate a VFS with a map[string]string
// (which is less noisy than map[string][]byte in test fixtures).
func mapFS(m map[string]string) ctxvfs.FileSystem {
	m2 := make(map[string][]byte, len(m))
	for k, v := range m {
		m2[k] = []byte(v)
	}
	return ctxvfs.Map(m2)
}
