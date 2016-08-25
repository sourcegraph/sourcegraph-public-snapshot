package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

var (
	mode    = flag.String("mode", "stdio", "communication mode (stdio|tcp)")
	addr    = flag.String("addr", ":2088", "server listen address (tcp)")
	logfile = flag.String("log", "/tmp/sample_server.log", "write log output to this file (and stderr)")
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	l := log.New(os.Stderr, "", 0)

	if *logfile != "" {
		f, err := os.Create(*logfile)
		if err != nil {
			return err
		}
		defer f.Close()
		l.SetOutput(f)
		log.SetOutput(f)
	}

	h := jsonrpc2.HandlerWithError(handle)

	switch *mode {
	case "tcp":
		lis, err := net.Listen("tcp", *addr)
		if err != nil {
			return err
		}
		defer lis.Close()
		l.Println("listening on", *addr)
		return jsonrpc2.Serve(context.Background(), lis, h, jsonrpc2.LogMessages(l))

	case "stdio":
		l.Println("reading on stdin, writing on stdout")
		<-jsonrpc2.NewConn(context.Background(), stdrwc{}, h, jsonrpc2.LogMessages(l)).DisconnectNotify()
		l.Println("connection closed")
		return nil

	default:
		return fmt.Errorf("invalid mode %q", *mode)
	}
}

type stdrwc struct{}

func (v stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (v stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}

func handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	switch req.Method {
	case "initialize":
		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				HoverProvider:      true,
				DefinitionProvider: true,
				ReferencesProvider: true,
			},
		}, nil

	case "shutdown":
		// Result is undefined, per
		// https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#shutdown-request.
		return nil, nil

	case "textDocument/hover":
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		pos := params.Position
		return lsp.Hover{
			Contents: []lsp.MarkedString{{Language: "markdown", Value: "Hello over LSP!"}},
			Range: lsp.Range{
				Start: lsp.Position{Line: pos.Line, Character: pos.Character - 3},
				End:   lsp.Position{Line: pos.Line, Character: pos.Character + 3},
			},
		}, nil

	case "textDocument/definition":
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return lsp.Location{
			URI: params.TextDocument.URI,
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 0, Character: 0},
			},
		}, nil

	case "textDocument/references":
		var params lsp.ReferenceParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return []lsp.Location{lsp.Location{
			URI: params.TextDocument.URI,
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 0, Character: 3},
			},
		}, lsp.Location{
			URI: params.TextDocument.URI,
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 4},
				End:   lsp.Position{Line: 0, Character: 6},
			},
		}}, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}
