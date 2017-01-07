package xlang

import (
	"context"
	"log"
	"net"
	"os"
	"strings"

	"github.com/sourcegraph/jsonrpc2"

	log15 "gopkg.in/inconshreveable/log15.v2"

	srccli "sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	gobuildserver "sourcegraph.com/sourcegraph/sourcegraph/xlang/golang/buildserver"
)

// This file contains helper commands and server hooks for development.

func init() {
	srccli.ServeInit = append(srccli.ServeInit, func() {
		if os.Getenv("LSP_PROXY") != "" {
			// To prevent mistakes, check if there are any explicitly
			// configured lang servers with `LANGSERVER_<lang>` env
			// vars. These would be ignored by this process, since it
			// will send all language requests directly to the LSP
			// proxy (which is the process on which the
			// `LANGSERVER_<lang>` env vars must be set).
			for _, kv := range os.Environ() {
				if strings.HasPrefix(kv, "LANGSERVER_") {
					log.Fatalf("Invalid configuration: The env var %q is mutually inconsistent with LSP_PROXY=%q. If you set an external LSP proxy, Sourcegraph will send all language requests directly to the LSP proxy, regardless of language/mode, so LANGSERVER_<lang> env vars on the Sourcegraph process would have no effect. The LANGSERVER_<lang> env vars must be set on the LSP proxy process.", kv, os.Getenv("LSP_PROXY"))
				}
			}
			return // don't start dev LSP proxy
		}

		// If LSP_PROXY is unset, then assume we're running in
		// development mode and start our own dev LSP proxy and dev
		// language servers.

		// Start and register a builtin Go language server if env var
		// LANGSERVER_GO is ":builtin:". This builtin Go language
		// server is special-cased to ensure that any Sourcegraph dev
		// server has support for at least one language (which makes
		// dev and testing easier), regardless of the execution
		// environment.
		if os.Getenv("LANGSERVER_GO") == ":builtin:" {
			l, err := net.Listen("tcp", ":0")
			if err != nil {
				log.Fatal("Builtin Go lang server: Listen:", err)
			}
			addr := "tcp://" + l.Addr().String()
			if err := os.Setenv("LANGSERVER_GO", addr); err != nil {
				log.Fatal("Set LANGSERVER_GO:", err)
			}
			if os.Getenv("LANGSERVER_GO_BG") == ":builtin:" {
				if err := os.Setenv("LANGSERVER_GO_BG", addr); err != nil {
					log.Fatal("Set LANGSERVER_GO_BG:", err)
				}
			}
			go func() {
				defer l.Close()
				for {
					conn, err := l.Accept()
					if err != nil {
						log.Fatal("Builtin Go lang server: Accept:", err)
					}
					jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(conn, jsonrpc2.VSCodeObjectCodec{}), gobuildserver.NewHandler())
				}
			}()
		}

		// Register language servers from `LANGSERVER_<lang>` env
		// vars (see RegisterServersFromEnv docs for more info.
		if err := RegisterServersFromEnv(); err != nil {
			log.Fatal(err)
		}

		// Start an in-process LSP proxy.
		l, err := net.Listen("tcp", ":4388")
		if err != nil {
			log.Fatal("LSP dev proxy: Listen:", err)
		}
		log15.Debug("Starting in-process LSP proxy.", "listen", l.Addr().String())
		proxy := NewProxy()
		if err := os.Setenv("LSP_PROXY", l.Addr().String()); err != nil {
			log.Fatal("Set LSP_PROXY:", err)
		}
		go func() {
			defer proxy.Close(context.Background())
			if err := proxy.Serve(context.Background(), l); err != nil {
				log.Fatal("LSP dev proxy:", err)
			}
		}()
	})
}
