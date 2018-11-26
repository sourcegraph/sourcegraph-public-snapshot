package server_test

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-lsp/lspext"
	"github.com/sourcegraph/jsonrpc2"
	gobuildserver "github.com/sourcegraph/sourcegraph/enterprise/cmd/xlang-go/internal/server"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
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
//   # Run this before making the changes that you want to benchmark:
//   go get -u golang.org/x/tools/cmd/benchcmp
//   go test github.com/sourcegraph/sourcegraph/xlang -bench=Integration -benchmem -run='^$' > /tmp/BenchmarkIntegration.old.txt
//
//   # Run this after you've made the changes that you want to benchmark:
//   go test github.com/sourcegraph/sourcegraph/xlang -bench=Integration -benchmem -run='^$' | tee /tmp/BenchmarkIntegration.new.txt && benchcmp /tmp/BenchmarkIntegration.{old,new}.txt
func BenchmarkIntegration(b *testing.B) {
	if testing.Short() {
		b.Skip("skip long integration test")
	}

	cleanup := useGithubForVFS()
	defer cleanup()

	tests := map[lsp.DocumentURI]struct { // map key is rootURI
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
	for rootURI, test := range tests {
		root, err := gituri.Parse(string(rootURI))
		if err != nil {
			b.Fatal(err)
		}
		label := strings.Replace(root.Host+root.Path, "/", "-", -1)

		b.Run(label, func(b *testing.B) {
			fs, err := gobuildserver.RemoteFS(context.Background(), lspext.InitializeParams{OriginalRootURI: rootURI})
			if err != nil {
				b.Fatal(err)
			}
			fs.Stat(context.Background(), ".") // ensure repo archive has been fetched and read before starting timer

			if test.definitionParams != nil {
				test.definitionParams.TextDocument.URI = rootURI + "#" + test.definitionParams.TextDocument.URI
				b.Run("definition", func(b *testing.B) {
					ctx := context.Background()

					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						b.StopTimer()
						c, done := connectionToNewBuildServer(string(rootURI), b)
						b.StartTimer()

						if err := c.Call(ctx, "initialize", lspext.ClientProxyInitializeParams{
							InitializeParams:      lsp.InitializeParams{RootURI: rootURI},
							InitializationOptions: lspext.ClientProxyInitializationOptions{Mode: test.mode},
						}, nil); err != nil {
							b.Fatal(err)
						}

						var loc locations
						if err := c.Call(ctx, "textDocument/definition", test.definitionParams, &loc); err != nil {
							b.Fatal(err)
						}

						// If we don't shut down the server, then subsequent
						// iterations will test the performance when it's
						// already cached, which is not what we want.
						b.StopTimer()
						done()
						if !reflect.DeepEqual(loc, test.wantDefinitions) {
							b.Fatalf("got %v, want %v", loc, test.wantDefinitions)
						}
						b.StartTimer()
					}
					b.StopTimer() // don't include server teardown
				})
			}

			if test.symbolParams != nil {
				b.Run("symbols", func(b *testing.B) {
					ctx := context.Background()

					for i := 0; i < b.N; i++ {
						b.StopTimer()
						c, done := connectionToNewBuildServer(string(rootURI), b)
						b.StartTimer()

						if err := c.Call(ctx, "initialize", lspext.ClientProxyInitializeParams{
							InitializeParams:      lsp.InitializeParams{RootURI: rootURI},
							InitializationOptions: lspext.ClientProxyInitializationOptions{Mode: test.mode},
						}, nil); err != nil {
							b.Fatal(err)
						}

						var syms []lsp.SymbolInformation
						if err := c.Call(ctx, "workspace/symbol", test.symbolParams, &syms); err != nil {
							b.Fatal(err)
						}

						// If we don't shut down the server, then subsequent
						// iterations will test the performance when it's
						// already cached, which is not what we want.
						b.StopTimer()
						done()
						if !reflect.DeepEqual(syms, test.wantSymbols) {
							b.Fatalf("got %v, want %v", syms, test.wantSymbols)
						}
						b.StartTimer()
					}
					b.StopTimer() // don't include server teardown
				})
			}
		})
	}
}

// BenchmarkIntegrationShared is a benchmark for testing how well we reuse
// previously computed artifacts.
//
// go test -c
// ./proxy.test -test.bench=IntegrationShared -test.v -test.run='^$
func BenchmarkIntegrationShared(b *testing.B) {
	if testing.Short() {
		b.Skip("skip long integration test")
	}

	cleanup := useGithubForVFS()
	defer cleanup()

	tests := map[string]struct {
		// oldRootURI will be run outside of the benchmark timers. It is where the reused artifacts will come from
		oldRootURI string

		rootURI string

		// Only a TextDocument is specified since we do the same
		// prepare hover as in production.
		path string
	}{
		// noop tests the case where no go files have changed between commits
		"noop": {
			oldRootURI: "git://github.com/kubernetes/kubernetes?ae03433a70ddb01b9c2be052a9ea0810395ff368",
			rootURI:    "git://github.com/kubernetes/kubernetes?c41c24fbf300cd7ba504ea1ac2e052c4a1bbed33",
			path:       "pkg/ssh/ssh.go",
		},
		// fast is a small change in one file
		"fast": {
			oldRootURI: "git://github.com/kubernetes/kubernetes?c41c24fbf300cd7ba504ea1ac2e052c4a1bbed33",
			rootURI:    "git://github.com/kubernetes/kubernetes?e105eec9c91afdd19e5245ddfadf2e2d2155eb6f",
			path:       "pkg/ssh/ssh.go",
		},
	}
	ctx := context.Background()
	for label, test := range tests {
		b.Run(label, func(b *testing.B) {
			oldfs, err := gobuildserver.RemoteFS(context.Background(), lspext.InitializeParams{OriginalRootURI: lsp.DocumentURI(test.oldRootURI)})
			if err != nil {
				b.Fatal(err)
			}
			fs, err := gobuildserver.RemoteFS(context.Background(), lspext.InitializeParams{OriginalRootURI: lsp.DocumentURI(test.rootURI)})
			if err != nil {
				b.Fatal(err)
			}
			// ensure repo archive has been fetched and read before starting timer
			oldfs.Stat(ctx, ".")
			fs.Stat(ctx, ".")

			do := func(c *jsonrpc2.Conn, rootURI string) {
				if err := c.Call(ctx, "initialize", lspext.ClientProxyInitializeParams{
					InitializeParams:      lsp.InitializeParams{RootURI: lsp.DocumentURI(rootURI)},
					InitializationOptions: lspext.ClientProxyInitializationOptions{Mode: "go"},
				}, nil); err != nil {
					b.Fatal(err)
				}

				var hover lsp.Hover
				if err := c.Call(ctx, "textDocument/hover", &lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{
						URI: lsp.DocumentURI(rootURI + "#" + test.path),
					},
				}, &hover); err != nil {
					b.Fatal(err)
				}

				if err := c.Close(); err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()

				c, done1 := connectionToNewBuildServer(test.oldRootURI, b)
				defer done1() // TODO ensure we close between each loop
				do(c, test.oldRootURI)
				c, done2 := connectionToNewBuildServer(test.rootURI, b)
				defer done2()

				b.StartTimer()
				do(c, test.rootURI)
				b.StopTimer()
			}
		})
	}
}
