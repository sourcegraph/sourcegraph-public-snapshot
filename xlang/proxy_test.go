package xlang_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	// Register Go server for testing.
	_ "sourcegraph.com/sourcegraph/sourcegraph/xlang/golang"

	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/mapfs"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspx"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

func TestProxy(t *testing.T) {
	t.Skip("Disabled https://github.com/sourcegraph/sourcegraph/issues/1331")
	tests := map[string]struct {
		rootPath       string
		mode           string
		fs             vfs.FileSystem
		wantHover      map[string]string
		wantDefinition map[string]string
		wantReferences map[string][]string
		otherVFS       map[string]vfs.FileSystem
	}{
		"go": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: mapfs.New(map[string]string{
				"a.go": "package p; func A() { A() }",
				"b.go": "package p; func B() { A() }",
			}),
			wantHover: map[string]string{
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
		},
		"go subdirectory in repo": {
			rootPath: "git://test/pkg?master#d",
			mode:     "go",
			fs: mapfs.New(map[string]string{
				"a.go":    "package d; func A() { A() }",
				"d2/b.go": `package d2; import "test/pkg/d"; func B() { d.A(); B() }`,
			}),
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
		},
		"go multiple packages in dir": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: mapfs.New(map[string]string{
				"a.go": "package p; func A() { A() }",
				"main.go": `// +build ignore

package main; import "test/pkg"; func B() { p.A(); B() }`,
			}),
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
		},
		"goroot": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: mapfs.New(map[string]string{
				"a.go": `package p; import "fmt"; var _ = fmt.Println`,
			}),
			wantHover: map[string]string{
				"a.go:1:40": "func Println(a ...interface{}) (n int, err error)",
			},
			wantDefinition: map[string]string{
				"a.go:1:40": "git://github.com/golang/go?" + runtime.Version() + "#src/fmt/print.go:1:19",
			},
			otherVFS: map[string]vfs.FileSystem{
				"https://github.com/golang/go?go1.7.1": mapfs.New(map[string]string{
					"src/fmt/print.go": "package fmt; func Println(a ...interface{}) (n int, err error) { return }",
				}),
			},
		},
		"gopath": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: mapfs.New(map[string]string{
				"a/a.go": `package a; func A() {}`,
				"b/b.go": `package b; import "test/pkg/a"; var _ = a.A`,
			}),
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
		},
		"go vendored dep": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: mapfs.New(map[string]string{
				"a.go": `package a; import "github.com/v/vendored"; var _ = vendored.V`,
				"vendor/github.com/v/vendored/v.go": "package vendored; func V() {}",
			}),
			wantHover: map[string]string{
				"a.go:1:61": "func V()",
			},
			wantDefinition: map[string]string{
				"a.go:1:61": "git://test/pkg?master#vendor/github.com/v/vendored/v.go:1:24",
			},
		},
		"go external dep": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: mapfs.New(map[string]string{
				"a.go": `package a; import "github.com/d/dep"; var _ = dep.D`,
			}),
			wantHover: map[string]string{
				"a.go:1:51": "func D()",
			},
			wantDefinition: map[string]string{
				"a.go:1:51": "git://github.com/d/dep?HEAD#d.go:1:19",
			},
			otherVFS: map[string]vfs.FileSystem{
				"https://github.com/d/dep?HEAD": mapfs.New(map[string]string{
					"d.go": "package dep; func D() {}",
				}),
			},
		},
		"external dep with vendor": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: mapfs.New(map[string]string{
				"a.go": `package p; import "github.com/d/dep"; var _ = dep.D().F`,
			}),
			wantDefinition: map[string]string{
				"a.go:1:55": "git://github.com/d/dep?HEAD#vendor/vendp/vp.go:1:32",
			},
			otherVFS: map[string]vfs.FileSystem{
				"https://github.com/d/dep?HEAD": mapfs.New(map[string]string{
					"d.go":               `package dep; import "vendp"; func D() (v vendp.V) { return }`,
					"vendor/vendp/vp.go": "package vendp; type V struct { F int }",
				}),
			},
		},
		"go external dep at subtree": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: mapfs.New(map[string]string{
				"a.go": `package a; import "github.com/d/dep/subp"; var _ = subp.D`,
			}),
			wantHover: map[string]string{
				"a.go:1:57": "func D()",
			},
			wantDefinition: map[string]string{
				"a.go:1:57": "git://github.com/d/dep?HEAD#subp/d.go:1:20",
			},
			otherVFS: map[string]vfs.FileSystem{
				"https://github.com/d/dep?HEAD": mapfs.New(map[string]string{
					"subp/d.go": "package subp; func D() {}",
				}),
			},
		},
		"go nested external dep": { // a depends on dep1, dep1 depends on dep2
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: mapfs.New(map[string]string{
				"a.go": `package a; import "github.com/d/dep1"; var _ = dep1.D1().D2`,
			}),
			wantHover: map[string]string{
				"a.go:1:53": "func D1() D2",
				"a.go:1:59": "D2 int",
			},
			wantDefinition: map[string]string{
				"a.go:1:53": "git://github.com/d/dep1?HEAD#d1.go:1:48", // func D1
				"a.go:1:58": "git://github.com/d/dep2?HEAD#d2.go:1:32", // field D2
			},
			otherVFS: map[string]vfs.FileSystem{
				"https://github.com/d/dep1?HEAD": mapfs.New(map[string]string{
					"d1.go": `package dep1; import "github.com/d/dep2"; func D1() dep2.D2 { return dep2.D2{} }`,
				}),
				"https://github.com/d/dep2?HEAD": mapfs.New(map[string]string{
					"d2.go": "package dep2; type D2 struct { D2 int }",
				}),
			},
		},
		"go external dep at vanity import path": {
			rootPath: "git://test/pkg?master",
			mode:     "go",
			fs: mapfs.New(map[string]string{
				"a.go": `package a; import "golang.org/x/text"; var _ = text.F`,
			}),
			wantHover: map[string]string{
				"a.go:1:53": "func F()",
			},
			wantDefinition: map[string]string{
				"a.go:1:53": "git://github.com/golang/text?HEAD#dummy.go:1:20",
			},
			otherVFS: map[string]vfs.FileSystem{
				// We override the Git cloning of this repo to use
				// in-memory dummy data, but we still need to hit the
				// network to resolve the Go custom import path
				// (because that's not mocked yet).
				"https://github.com/golang/text?HEAD": mapfs.New(map[string]string{
					"dummy.go": "package text; func F() {}",
				}),
			},
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			orig := xlang.VFSCreatorsByScheme["git"]
			xlang.VFSCreatorsByScheme["git"] = func(root *uri.URI) (vfs.FileSystem, error) {
				if fs, ok := test.otherVFS[root.String()]; ok {
					return fs, nil
				}
				return test.fs, nil
			}
			defer func() {
				xlang.VFSCreatorsByScheme["git"] = orig
			}()

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
	t.Skip("Disabled https://github.com/sourcegraph/sourcegraph/issues/1331")
	ctx := context.Background()

	xlang.VFSCreatorsByScheme["test"] = func(root *uri.URI) (vfs.FileSystem, error) {
		return mapfs.New(map[string]string{"f": "x"}), nil
	}
	defer func() {
		delete(xlang.VFSCreatorsByScheme, "test")
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
		InitializeParams: lsp.InitializeParams{RootPath: "test://test"},
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
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://test#myfile"},
		Position:     lsp.Position{Line: 1, Character: 2},
	}, nil); err != nil {
		t.Fatal(err)
	}
	want := []testRequest{
		{"initialize", lspx.InitializeParams{
			InitializeParams: lsp.InitializeParams{RootPath: "file:///"},
			OriginalRootPath: "test://test"}},
		{"textDocument/didOpen", lsp.DidOpenTextDocumentParams{
			TextDocument: lsp.TextDocumentItem{URI: "file:///f", Text: "x"}}},
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
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://test#myfile2"},
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
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://test#myfile3"},
		Position:     lsp.Position{Line: 5, Character: 6},
	}, nil); err != nil {
		t.Fatal(err)
	}
	want = []testRequest{
		{"shutdown", nil},
		{"exit", nil},
		{"initialize", lspx.InitializeParams{
			InitializeParams: lsp.InitializeParams{RootPath: "file:///"},
			OriginalRootPath: "test://test"}},
		{"textDocument/didOpen", lsp.DidOpenTextDocumentParams{
			TextDocument: lsp.TextDocumentItem{URI: "file:///f", Text: "x"}}},
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

	xlang.VFSCreatorsByScheme["test"] = func(root *uri.URI) (vfs.FileSystem, error) {
		return mapfs.New(map[string]string{"f": "x"}), nil
	}
	defer func() {
		delete(xlang.VFSCreatorsByScheme, "test")
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
		InitializeParams: lsp.InitializeParams{RootPath: "test://test"},
		Mode:             "test",
	}
	if err := c.Call(ctx, "initialize", initParams, nil); err != nil {
		t.Fatal(err)
	}

	// Call something that triggers the server to return diagnostics.
	if err := c.Call(ctx, "textDocument/definition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://test#myfile"},
		Position:     lsp.Position{Line: 1, Character: 2},
	}, nil); err != nil {
		t.Fatal(err)
	}

	// Check that we got the diagnostics.
	select {
	case diags := <-recvDiags:
		want := lsp.PublishDiagnosticsParams{
			URI: "test://test#myfile",
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

type testRequests []testRequest

func (v testRequests) Len() int      { return len(v) }
func (v testRequests) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v testRequests) Less(i, j int) bool {
	ii, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	jj, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	return string(ii) < string(jj)
}
