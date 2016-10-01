package xlang_test

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/ctxvfs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

// TestServerProxy_sendFiles tests that the server proxy sends files
// and waits for an ack before sending other requests.
func TestServerProxy_sendFiles(t *testing.T) {
	ctx := context.Background()

	xlang.VFSCreatorsByScheme["test"] = func(root *uri.URI) (ctxvfs.FileSystem, error) {
		return ctxvfs.Map(map[string][]byte{"f": []byte("x")}), nil
	}
	defer func() {
		delete(xlang.VFSCreatorsByScheme, "test")
	}()

	// Start test build/lang server that has an artificial delay
	// receiving files to simulate racy conditions. We test below that
	// the client waits for the textDocument/didOpen ack before it
	// sends the textDocument/hover request.
	var mu sync.Mutex
	receivedFile := false
	xlang.ServersByMode["test"] = func() (io.ReadWriteCloser, error) {
		a, b := xlang.InMemoryPeerConns()
		jsonrpc2.NewConn(context.Background(), a, jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
			switch req.Method {
			case "textDocument/didOpen":
				time.Sleep(50 * time.Millisecond)
				mu.Lock()
				receivedFile = true
				mu.Unlock()
				return nil, nil

			case "textDocument/hover":
				mu.Lock()
				receivedFile := receivedFile
				mu.Unlock()
				if !receivedFile {
					return nil, errors.New("textDocument/hover called before file was received")
				}
				return nil, nil

			case "initialize", "shutdown", "exit":
				return nil, nil

			default:
				panic("unexpected method: " + req.Method)
			}
		}))
		return b, nil
	}
	defer func() {
		delete(xlang.ServersByMode, "test")
	}()

	proxy := xlang.NewProxy()
	addr, done := startProxy(t, proxy)
	defer done()

	// C connects to the proxy.
	c := dialProxy(t, addr, nil)
	initParams := xlang.ClientProxyInitializeParams{
		InitializeParams: lsp.InitializeParams{RootPath: "test://test"},
		Mode:             "test",
	}
	if err := c.Call(ctx, "initialize", initParams, nil); err != nil {
		t.Fatal(err)
	}

	// Now C sends an actual request. The proxy should open a
	// connection to S, send the file, and then send the request (*after*
	// the file has been acked).
	if err := c.Call(ctx, "textDocument/hover", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://test#myfile"},
		Position:     lsp.Position{Line: 1, Character: 2},
	}, nil); err != nil {
		t.Fatal(err)
	}

}
