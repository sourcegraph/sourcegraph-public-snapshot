package xlang_test

import (
	"context"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

// Notable benchmark results:
//
// When the build server passes ALL files to the lang server using textDocument/didOpen:
//
//   BenchmarkIntegration/github.com-gorilla-mux-8	1	1984109426 ns/op	612749040 B/op	 3017725 allocs/op
//
// When the build server shares an in-memory VFS with the lang server:
//
//   BenchmarkIntegration/github.com-gorilla-mux-8	2	 615687151 ns/op	325774544 B/op	 3255045 allocs/op
//
// When no files are present and the build server accesses files over a VFS residing on the LSP proxy:
//
//   BenchmarkIntegration/github.com-gorilla-mux-8	2	 722473764 ns/op	309509740 B/op	 3114545 allocs/op
//
// As of ed4362e65f1f7ffa643d5af8faf63ceaaef979b0:
//
//   BenchmarkIntegration/github.com-golang-go-definition-12  1  4332188655 ns/op  689994240 B/op  2166801 allocs/op
//
// To compare old vs. new benchmark results:
//
//   go get -u golang.org/x/tools/cmd/benchcmp
//   go test sourcegraph.com/sourcegraph/sourcegraph/xlang -bench=Integration -benchmem -run='^$' | tee /tmp/BenchmarkIntegration.new.txt && benchcmp /tmp/BenchmarkIntegration.{old,new}.txt
func BenchmarkIntegration(b *testing.B) {
	if testing.Short() {
		b.Skip("skip long integration test")
	}

	{
		// Serve repository data from codeload.github.com for
		// test performance instead of from gitserver. This
		// technically means we aren't testing gitserver, but
		// that is well tested separately, and the benefit of
		// fast tests here outweighs the benefits of a coarser
		// integration test.
		orig := xlang.NewRemoteRepoVFS
		xlang.NewRemoteRepoVFS = func(cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
			fullName := cloneURL.Host + strings.TrimSuffix(cloneURL.Path, ".git") // of the form "github.com/foo/bar"
			return vfsutil.NewGitHubRepoVFS(fullName, rev, "", true)
		}
		defer func() {
			xlang.NewRemoteRepoVFS = orig
		}()
	}

	tests := map[string]struct { // map key is rootPath
		mode string

		definitionParams *lsp.TextDocumentPositionParams
		wantDefinitions  locations

		symbolParams *lsp.WorkspaceSymbolParams
		wantSymbols  []lsp.SymbolInformation
	}{
		"git://github.com/gorilla/mux?0a192a193177452756c362c20087ddafcf6829c4": {
			mode: "go",
			definitionParams: &lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: "mux.go"},
				Position:     lsp.Position{Line: 60, Character: 37},
			},
			wantDefinitions: locations{
				{
					URI: "git://github.com/golang/go?go1.7.1#src/net/http/request.go",
					Range: lsp.Range{
						Start: lsp.Position{Line: 75, Character: 5},
						End:   lsp.Position{Line: 75, Character: 12},
					},
				},
			},
		},
		"git://github.com/golang/go?go1.7.1": {
			mode: "go",
			definitionParams: &lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: "src/fmt/print.go"},
				Position:     lsp.Position{Line: 189, Character: 12}, // "Fprintf" call
			},
			wantDefinitions: locations{
				{
					URI: "git://github.com/golang/go?go1.7.1#src/fmt/print.go",
					Range: lsp.Range{
						Start: lsp.Position{Line: 178, Character: 5},
						End:   lsp.Position{Line: 178, Character: 12},
					},
				},
			},
			symbolParams: &lsp.WorkspaceSymbolParams{Query: "NewUnstartedServer"},
			wantSymbols: []lsp.SymbolInformation{
				{
					ContainerName: "httptest",
					Name:          "NewUnstartedServer",
					Kind:          lsp.SKFunction,
					Location: lsp.Location{
						URI: "git://github.com/golang/go?go1.7.1#src/net/http/httptest/server.go",
						Range: lsp.Range{
							Start: lsp.Position{Line: 84, Character: 5},
							End:   lsp.Position{Line: 84, Character: 22},
						},
					},
				},
			},
		},
	}
	for rootPath, test := range tests {
		root, err := url.Parse(rootPath)
		if err != nil {
			b.Fatal(err)
		}
		label := strings.Replace(root.Host+root.Path, "/", "-", -1)

		if test.definitionParams != nil {
			test.definitionParams.TextDocument.URI = rootPath + "#" + test.definitionParams.TextDocument.URI
			b.Run(label+"-definition", func(b *testing.B) {
				ctx := context.Background()
				proxy := xlang.NewProxy()
				addr, done := startProxy(b, proxy)
				defer done()

				for i := 0; i < b.N; i++ {
					b.StopTimer()
					c := dialProxy(b, addr, nil)
					b.StartTimer()

					if err := c.Call(ctx, "initialize", xlang.ClientProxyInitializeParams{
						InitializeParams: lsp.InitializeParams{RootPath: rootPath},
						Mode:             test.mode,
					}, nil); err != nil {
						b.Fatal(err)
					}

					var loc locations
					if err := c.Call(ctx, "textDocument/definition", test.definitionParams, &loc); err != nil {
						b.Fatal(err)
					}

					if err := c.Close(); err != nil {
						b.Fatal(err)
					}

					// If we don't shut down the server, then subsequent
					// iterations will test the performance when it's
					// already cached, which is not what we want.
					b.StopTimer()
					proxy.ShutDownIdleServers(ctx, 0)
					if !reflect.DeepEqual(loc, test.wantDefinitions) {
						b.Fatalf("got %v, want %v", loc, test.wantDefinitions)
					}
					b.StartTimer()
				}
				b.StopTimer() // don't include server teardown
			})
		}

		if test.symbolParams != nil {
			b.Run(label+"-symbols", func(b *testing.B) {
				ctx := context.Background()
				proxy := xlang.NewProxy()
				addr, done := startProxy(b, proxy)
				defer done()

				for i := 0; i < b.N; i++ {
					b.StopTimer()
					c := dialProxy(b, addr, nil)
					b.StartTimer()

					if err := c.Call(ctx, "initialize", xlang.ClientProxyInitializeParams{
						InitializeParams: lsp.InitializeParams{RootPath: rootPath},
						Mode:             test.mode,
					}, nil); err != nil {
						b.Fatal(err)
					}

					var syms []lsp.SymbolInformation
					if err := c.Call(ctx, "workspace/symbol", test.symbolParams, &syms); err != nil {
						b.Fatal(err)
					}

					if err := c.Close(); err != nil {
						b.Fatal(err)
					}

					// If we don't shut down the server, then subsequent
					// iterations will test the performance when it's
					// already cached, which is not what we want.
					b.StopTimer()
					proxy.ShutDownIdleServers(ctx, 0)
					if !reflect.DeepEqual(syms, test.wantSymbols) {
						b.Fatalf("got %v, want %v", syms, test.wantSymbols)
					}
					b.StartTimer()
				}
				b.StopTimer() // don't include server teardown
			})
		}
	}
}
