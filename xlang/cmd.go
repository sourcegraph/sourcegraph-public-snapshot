package xlang

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
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
			if err := os.Setenv("LANGSERVER_GO", "tcp://"+l.Addr().String()); err != nil {
				log.Fatal("Set LANGSERVER_GO:", err)
			}
			go func() {
				defer l.Close()
				for {
					conn, err := l.Accept()
					if err != nil {
						log.Fatal("Builtin Go lang server: Accept:", err)
					}
					jsonrpc2.NewConn(context.Background(), conn, gobuildserver.NewHandler())
				}
			}()
		}

		// Register language servers from `LANGSERVER_<lang>` env
		// vars (see RegisterServersFromEnv docs for more info.
		if err := RegisterServersFromEnv(); err != nil {
			log.Fatal(err)
		}

		// Start an in-process LSP proxy.
		l, err := net.Listen("tcp", ":0")
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

func init() {
	c, err := srccli.CLI.AddCommand("xlang", "xlang", "", &cmd{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("repro", "reproduce a tracked error", "", &reproCmd{})
	if err != nil {
		log.Fatal(err)
	}
}

type cmd struct{}

func (cmd) Execute(args []string) error { return nil }

type reproCmd struct {
	BaseURL string `long:"base-url" description:"base URL of Sourcegraph site to repro request against" default:"https://sourcegraph.com"`
}

func (c *reproCmd) Execute(args []string) error {
	t := time.AfterFunc(250*time.Millisecond, func() {
		fmt.Fprintln(os.Stderr, "Waiting for JSON input on stdin...")
	})

	var terr trackedError
	if err := json.NewDecoder(os.Stdin).Decode(&terr); err != nil {
		return err
	}
	t.Stop()

	u, err := guessTrackedErrorURL(terr)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Repro URLs:")
	fmt.Fprintln(os.Stderr, " - https://sourcegraph.com/"+u)
	fmt.Fprintln(os.Stderr, " - http://localhost:3080/"+u)
	fmt.Fprintln(os.Stderr)

	lspReqs := []jsonrpc2.Request{
		{ID: 1, Method: "initialize"},
		{ID: 2, Method: terr.Method},
		{ID: 3, Method: "shutdown"},
		{Notif: true, Method: "exit"},
	}
	if err := lspReqs[0].SetParams(lsp.InitializeParams{RootPath: terr.RootPath}); err != nil {
		return err
	}
	if err := lspReqs[1].SetParams(terr.Params); err != nil {
		return err
	}
	body, err := json.MarshalIndent(lspReqs, "   ", "  ")
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/.api/xlang/%s", c.BaseURL, terr.Method), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json; charset=utf-8")

	fmt.Fprintf(os.Stderr, "<<< Sending HTTP POST to %s with request body:\n", req.URL)
	fmt.Fprintln(os.Stderr, "   "+string(body))
	fmt.Fprintln(os.Stderr)

	t0 := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		var out bytes.Buffer
		if err := json.Indent(&out, respBody, "  ", "   "); err != nil {
			return fmt.Errorf("%s (response body follows)\n\n%s", err, string(respBody))
		}
		respBody = out.Bytes()
	}
	if len(bytes.TrimSpace(respBody)) == 0 {
		respBody = []byte("(empty body)")
	}
	fmt.Fprintf(os.Stderr, ">>> Received HTTP %s with response body:\n", resp.Status)
	fmt.Fprintln(os.Stderr, "   "+string(respBody))
	fmt.Fprintln(os.Stderr)

	fmt.Fprintln(os.Stderr, " - Roundtrip:", time.Since(t0))
	fmt.Fprintln(os.Stderr, " - Lightstep trace:", resp.Header.Get("x-trace"))

	return nil
}

func guessTrackedErrorURL(lspReq interface{}) (string, error) {
	b, err := json.Marshal(lspReq)
	if err != nil {
		return "", err
	}
	var p struct {
		Params *lsp.TextDocumentPositionParams
	}
	if err := json.Unmarshal(b, &p); err != nil || p.Params == nil || p.Params.TextDocument.URI == "" {
		return "", err
	}
	u, err := url.Parse(p.Params.TextDocument.URI)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s@%s/-/blob/%s#L%d:%d", path.Join(u.Host, u.Path), u.RawQuery, u.Fragment, p.Params.Position.Line+1, p.Params.Position.Character+1), nil
}
