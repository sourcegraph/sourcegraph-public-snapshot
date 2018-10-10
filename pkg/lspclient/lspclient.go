package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

var (
	url         = flag.String("url", "ws://localhost:3080/.api/lsp", "WebSocket LSP gateway URL (ws:// or wss://)") // CI:LOCALHOST_OK
	accessToken = flag.String("token", "", "Sourcegraph API access token")
	mode        = flag.String("mode", "go", "")
	rootURI     = flag.String("root-uri", "git://github.com/sourcegraph/go-diff?3f415a150aec0685cb81b73cc201e762e075006d", "LSP initialize rootURI (git://repo?sha)")
	path        = flag.String("path", "diff/diff.go", "textDocument/hover textDocument path (appended to root URI)")
	line        = flag.Int("line", 20, "line number (0-indexed)")
	char        = flag.Int("char", 15, "character (0-indexed)")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	if *accessToken == "" {
		log.Fatalf("Must set -token flag to Sourcegraph API access token valid for %s", *url)
	}

	// Create unique session ID for isolation.
	session := make([]byte, 40)
	if _, err := rand.Read(session); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	h := http.Header{}
	h.Set("Authorization", "token "+*accessToken)
	websocket.DefaultDialer.HandshakeTimeout = 500 * time.Millisecond
	wc, resp, err := websocket.DefaultDialer.Dial(*url, h)
	if err != nil {
		if resp != nil {
			body, _ := ioutil.ReadAll(resp.Body)
			if max := 500; len(body) > max {
				body = body[:max]
			}
			log.Printf("WebSocket dial returned HTTP %d %s (%s)", resp.StatusCode, http.StatusText(resp.StatusCode), err)
			if len(body) > 0 {
				log.Println()
				log.Println(string(body))
			}
			os.Exit(1)
		} else {
			log.Fatal(err)
		}
	}
	log.Printf("# Connected to WebSocket LSP gateway at %s", *url)
	log.Println()

	c := jsonrpc2.NewConn(ctx, websocketjsonrpc2.NewObjectStream(wc), jsonrpc2HandlerFunc(handler))
	defer c.Close()
	go func() {
		<-c.DisconnectNotify()
		log.Fatal("# WebSocket LSP gateway disconnected")
	}()

	var result json.RawMessage
	if err := c.Call(ctx, "initialize", &lspext.ClientProxyInitializeParams{
		InitializeParams: lsp.InitializeParams{
			RootURI: lsp.DocumentURI(*rootURI),
		},
		InitializationOptions: lspext.ClientProxyInitializationOptions{
			Mode:    *mode,
			Session: string(session),
		},
	}, &result); err != nil {
		log.Fatal(err)
	}
	log.Println("# initialize response:")
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
	log.Println()

	if err := c.Call(ctx, "textDocument/hover", &lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: lsp.DocumentURI(*rootURI + "#" + *path)},
		Position:     lsp.Position{Line: *line, Character: *char},
	}, &result); err != nil {
		log.Fatal(err)
	}
	log.Println("# textDocument/hover response:")
	b, err = json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
	log.Println()
}

func handler(ctx context.Context, c *jsonrpc2.Conn, r *jsonrpc2.Request) {
	log.Println("Request from server:", r.Method)
}

type jsonrpc2HandlerFunc func(context.Context, *jsonrpc2.Conn, *jsonrpc2.Request)

func (h jsonrpc2HandlerFunc) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	h(ctx, conn, req)
}
