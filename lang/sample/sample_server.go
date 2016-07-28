package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
	if *logfile != "" {
		f, err := os.Create(*logfile)
		if err != nil {
			return err
		}
		defer f.Close()
		log.SetOutput(io.MultiWriter(os.Stderr, f))
	}

	h := &jsonrpc2.LoggingHandler{handler{}}

	switch *mode {
	case "tcp":
		lis, err := net.Listen("tcp", *addr)
		if err != nil {
			return err
		}
		defer lis.Close()
		log.Println("listening on", *addr)
		return jsonrpc2.Serve(lis, h)

	case "stdio":
		log.Println("reading on stdin, writing on stdout")
		jsonrpc2.NewServerConn(os.Stdin, os.Stdout, h)
		select {}

	default:
		return fmt.Errorf("invalid mode %q", *mode)
	}
}

type handler struct{}

func (handler) Handle(req *jsonrpc2.Request) (resp *jsonrpc2.Response) {
	if !req.Notification {
		resp = &jsonrpc2.Response{ID: req.ID}
	}

	switch req.Method {
	case "initialize":
		resp.SetResult(lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				HoverProvider:      true,
				DefinitionProvider: true,
				ReferencesProvider: true,
			},
		})

	case "shutdown":
		// Result is undefined, per
		// https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#shutdown-request.
		resp.SetResult(true)

	case "textDocument/hover":
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			resp.Error = &jsonrpc2.Error{Code: 123, Message: "error!"}
			return
		}

		pos := params.Position
		resp.SetResult(lsp.Hover{
			Contents: []lsp.MarkedString{{Language: "markdown", Value: "Hello over LSP!"}},
			Range: lsp.Range{
				Start: lsp.Position{Line: pos.Line, Character: pos.Character - 3},
				End:   lsp.Position{Line: pos.Line, Character: pos.Character + 3},
			},
		})
	case "textDocument/definition":
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			resp.Error = &jsonrpc2.Error{Code: 123, Message: "error!"}
			return
		}
		resp.SetResult(lsp.Location{
			URI: params.TextDocument.URI,
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 0, Character: 0},
			},
		})
	case "textDocument/references":
		var params lsp.ReferenceParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			resp.Error = &jsonrpc2.Error{Code: 123, Message: "error!"}
			return
		}
		resp.SetResult([]lsp.Location{lsp.Location{
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
		}})
	}

	return
}
