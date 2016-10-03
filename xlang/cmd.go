package xlang

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"

	log15 "gopkg.in/inconshreveable/log15.v2"

	srccli "sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/ctxvfs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

func init() {
	c, err := srccli.CLI.AddCommand("xlang", "xlang", "", &cmd{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("send", "send an LSP request to a build/lang server", "", &sendCmd{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("sendx", "send an LSP request to a build/lang server based on tracked error output.", "", &sendCmdSimple{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("url", "guess URL for where a user clicked based on tracked error output.", "", &urlCmd{})
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	srccli.ServeInit = append(srccli.ServeInit, func() {
		// If we're running "src serve" and haven't specified an
		// LSP_PROXY, then we're probably in development mode. Start
		// an in-process LSP proxy for convenience.
		if addr := os.Getenv("LSP_PROXY"); addr == "" {
			if err := RegisterServersFromEnv(); err != nil {
				log.Fatal(err)
			}
			addr, run, _, err := devProxy()
			if err != nil {
				log.Fatal("LSP dev proxy:", err)
			}
			log15.Debug("Starting in-process LSP proxy.", "listen", addr)
			if err := os.Setenv("LSP_PROXY", addr); err != nil {
				log.Fatal("Set LSP_PROXY:", err)
			}
			go func() {
				if err := run(); err != nil {
					log.Fatal("LSP dev proxy:", err)
				}
			}()
		} else {
			log15.Info("Using LSP proxy.", "addr", addr)
		}
	})
}

func devProxy() (addr string, run, done func() error, err error) {
	// The LSP dev proxy also assumes that a gitserver is available in
	// gitserver.DefaultClient. If devProxy is called outside of `src
	// serve` (which handles that already), you will need to connect
	// to gitservers explicitly, or override how repos are fetched in
	// NewRemoteRepoVFS.
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return
	}
	addr = l.Addr().String()
	proxy := NewProxy()
	run = func() error {
		return proxy.Serve(context.Background(), l)
	}
	done = func() error {
		return proxy.Close(context.Background())
	}
	return
}

type cmd struct{}

func (c *cmd) Execute(args []string) error {
	return nil
}

type sendCmdOpts struct {
	Addr            string        `long:"addr" description:"LSP proxy server address" env:"LSP_PROXY"`
	Trace           bool          `long:"trace" description:"print traces"`
	DiagnosticsWait time.Duration `long:"diagnostics-wait" description:"wait for the server to publish diagnostics asynchronously" default:"200ms"`
	Quiet           bool          `short:"q" long:"quiet" description:"print minimal output (only JSON result)"`
}

type sendCmdArgs struct {
	RootPath string `name:"ROOT-PATH" description:"rootPath for LSP initialization"`
	Mode     string `name:"MODE" description:"mode ID ('go', 'javascript', 'typescript', etc.)"`
	Method   string `name:"LSP-METHOD" description:"name of LSP method to send (e.g., textDocument/hover)"`
}

type sendCmd struct {
	sendCmdOpts
	Args sendCmdArgs `positional-args:"yes" required:"yes" count:"1"`
}

type sendCmdSimple struct {
	sendCmdOpts
	Input string `long:"input" description:"JSON blob describing Args and params for sendCmd"`
}

/*

Useful one-liners:

# Use in-process dev LSP proxy.
export LSP_PROXY=:dev:

echo '{"textDocument":{"uri":"git://github.com/gorilla/mux?0a192a1#mux.go"},"position":{"line":60,"character":37}}' | src xlang send git://github.com/gorilla/mux?0a192a1 go textDocument/definition

echo '{"textDocument":{"uri":"git://github.com/gorilla/mux?0a192a1#mux.go"},"position":{"line":60,"character":37}}' | src xlang send git://github.com/gorilla/mux?0a192a1 go textDocument/hover

echo '{"textDocument":{"uri":"git://github.com/gorilla/websocket?2d1e4548#client.go"},"position":{"line":23,"character":15}}' | src xlang send git://github.com/gorilla/websocket?2d1e4548 go textDocument/hover

echo '{"textDocument":{"uri":"git://github.com/golang/go?6129f3736#src/io/io.go"},"position":{"line":131,"character":12}}' | src xlang send git://github.com/golang/go?6129f3736 go textDocument/hover

echo '{"textDocument":{"uri":"git://github.com/golang/go?go1.7.1#src/net/http/client.go"},"position":{"line":134,"character":22}}' | src xlang send git://github.com/golang/go?go1.7.1 go textDocument/hover

echo '{"textDocument":{"uri":"git://github.com/docker/machine?e1a03348#libmachine/provision/provisioner.go"},"position":{"line":106,"character":49}}' | src xlang send git://github.com/docker/machine?e1a03348 go textDocument/hover

# Needs lots of external deps (they're vendored, but in a non-standard way in GOPATH=./vendor).
echo '{"textDocument":{"uri":"git://github.com/docker/docker?b16bfbad#daemon/create.go"},"position":{"line":110,"character":41}}' | src xlang send git://github.com/docker/docker?b16bfbad go textDocument/hover

echo '{"textDocument":{"uri":"git://github.com/docker/docker?762556c#pkg/symlink/fs.go"},"position":{"line":62,"character":5}}' | src xlang send git://github.com/docker/docker?762556c go textDocument/hover

# k8s is a huge repo
echo '{"textDocument":{"uri":"git://github.com/kubernetes/kubernetes?2580157#pkg/controller/informers/extensions.go"},"position":{"line":54,"character":26}}' | src xlang send git://github.com/kubernetes/kubernetes?2580157 go textDocument/hover

*/

func (c *sendCmd) Execute(args []string) error {
	reqParams, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	return sendExecute(&c.sendCmdOpts, &c.Args, reqParams)
}

func (c *sendCmdSimple) Execute(args []string) error {
	input := &struct {
		sendCmdArgs
		Params *json.RawMessage
	}{}
	err := json.Unmarshal([]byte(c.Input), input)
	if err != nil {
		return err
	}
	reqParams, _ := input.Params.MarshalJSON()
	logPrefix := []byte("tracked error: ")
	if bytes.HasPrefix(reqParams, logPrefix) {
		reqParams = reqParams[len(logPrefix):]
	}
	return sendExecute(&c.sendCmdOpts, &input.sendCmdArgs, reqParams)
}

func sendExecute(c *sendCmdOpts, args *sendCmdArgs, reqParams []byte) error {
	printIndentJSON := func(v json.RawMessage) {
		var buf bytes.Buffer
		if err := json.Indent(&buf, v, "", "  "); err == nil {
			fmt.Println(buf.String())
		} else {
			log.Println(err)
		}
	}

	// Fetch code from codeload.github.com, not from gitserver (which
	// is not available in `src xlang sendx?` because nowhere do we
	// spin it up or connect to it.)
	NewRemoteRepoVFS = func(cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
		fullName := cloneURL.Host + strings.TrimSuffix(cloneURL.Path, ".git") // of the form "github.com/foo/bar"
		return vfsutil.NewGitHubRepoVFS(fullName, rev, "", true)
	}
	fmt.Fprintln(os.Stderr, "# Fetching code from codeload.github.com, not gitserver")

	if c.Addr == ":dev:" {
		if t := os.Getenv("LIGHTSTEP_ACCESS_TOKEN"); t != "" {
			opentracing.InitGlobalTracer(lightstep.NewTracer(lightstep.Options{
				AccessToken: t,
			}))
		}

		addr, run, done, err := devProxy()
		if err != nil {
			return err
		}
		c.Addr = addr
		log.Println("# using in-process LSP proxy")
		defer done()
		go func() {
			if err := run(); err != nil {
				log.Fatal("LSP dev proxy:", err)
			}
		}()
	}
	if c.Addr == "" {
		return errors.New("must specify LSP proxy address in --addr option or LSP_PROXY env var")
	}

	if c.Quiet && c.Trace {
		return errors.New("options -q/--quiet and --trace are mutually exclusive")
	}

	var connOpt []jsonrpc2.ConnOpt
	if c.Trace {
		connOpt = append(connOpt, jsonrpc2.LogMessages(log.New(os.Stderr, "", 0)))
	}

	h := &ClientHandler{
		RecvDiagnostics: func(uri string, diags []lsp.Diagnostic) {
			if c.Quiet {
				return
			}
			fmt.Fprintf(os.Stderr, "# received diagnostics for %s:\n", uri)
			for _, d := range diags {
				fmt.Fprintf(os.Stderr, "#   :%d:%d: %s\n", d.Range.Start.Line+1, d.Range.Start.Character+1, d.Message)
			}
		},
	}

	dialCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	jc, err := DialProxy(dialCtx, c.Addr, h, connOpt...)
	if err != nil {
		return err
	}
	defer jc.Close()
	ctx := context.Background()

	var initResult lsp.InitializeResult
	if err := jc.Call(ctx, "initialize", ClientProxyInitializeParams{
		Mode:             args.Mode,
		InitializeParams: lsp.InitializeParams{RootPath: args.RootPath},
	}, &initResult); err != nil {
		return err
	}

	var result *json.RawMessage
	tReq0 := time.Now()
	if err := jc.Call(ctx, args.Method, (*json.RawMessage)(&reqParams), &result); err != nil {
		return err
	}
	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "# %s took %s\n", args.Method, time.Since(tReq0))
	}
	printIndentJSON(*result)

	time.Sleep(c.DiagnosticsWait)

	return nil
}

type urlCmd struct {
	Input string `long:"input" description:"JSON blob describing from tracked error logs"`
}

func (c *urlCmd) Execute(args []string) error {
	u := guessTrackedErrorURL([]byte(c.Input))
	if u == "" {
		return nil
	}
	log.Println("# The URL is a guess. It should be correct for standard simple repos.")
	fmt.Println("https://sourcegraph.com/" + u)
	fmt.Println("http://localhost:3080/" + u)
	return nil
}

func guessTrackedErrorURL(input []byte) string {
	t := &struct {
		Params *lsp.TextDocumentPositionParams
	}{}
	err := json.Unmarshal(input, t)
	if err != nil || t.Params == nil {
		return ""
	}
	p := t.Params
	u, err := url.Parse(p.TextDocument.URI)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s@%s/-/blob/%s#L%d:%d", path.Join(u.Host, u.Path), u.RawQuery, u.Fragment, p.Position.Line+1, p.Position.Character+1)
}
