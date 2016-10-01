package xlang_test

import (
	"context"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
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
func BenchmarkIntegration(b *testing.B) {
	if testing.Short() {
		b.Skip("skip long integration test")
	}

	tests := map[string]struct { // map key is rootPath
		mode   string
		params lsp.TextDocumentPositionParams
		want   locations
	}{
		"git://github.com/gorilla/mux?0a192a193177452756c362c20087ddafcf6829c4": {
			mode: "go",
			params: lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: "mux.go"},
				Position:     lsp.Position{Line: 60, Character: 37},
			},
			want: locations{
				{
					URI: "git://github.com/golang/go?go1.7.1#src/net/http/request.go",
					Range: lsp.Range{
						Start: lsp.Position{Line: 75, Character: 5},
						End:   lsp.Position{Line: 75, Character: 12},
					},
				},
			},
		},
	}
	for rootPath, test := range tests {
		u, err := url.Parse(rootPath)
		if err != nil {
			b.Fatal(err)
		}
		test.params.TextDocument.URI = rootPath + "#" + test.params.TextDocument.URI
		label := strings.Replace(u.Host+u.Path, "/", "-", -1)

		b.Run(label, func(b *testing.B) {
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
				if err := c.Call(ctx, "textDocument/definition", test.params, &loc); err != nil {
					b.Fatal(err)
				}

				if err := c.Close(); err != nil {
					b.Fatal(err)
				}

				if !reflect.DeepEqual(loc, test.want) {
					b.Fatalf("got %v, want %v", loc, test.want)
				}

				// If we don't shut down the server, then subsequent
				// iterations will test the performance when it's
				// already cached, which is not what we want.
				b.StopTimer()
				proxy.ShutDownIdleServers(ctx, 0)
				b.StartTimer()
			}
			b.StopTimer() // don't include server teardown
		})
	}
}
