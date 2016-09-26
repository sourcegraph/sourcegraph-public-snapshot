package golang

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	pathpkg "path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspx"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

// NewBuildHandler creates a new build server wrapping a (also newly
// created) Go language server. I.e., it creates a BuildHandler
// wrapping a LangHandler. The two handlers share a file system (in
// memory).
//
// The build server is responsible for things such as fetching
// dependencies, setting up the right file system structure and paths,
// and mapping local file system paths to logical URIs (e.g.,
// /goroot/src/fmt/print.go ->
// git://github.com/golang/go?go1.7.1#src/fmt/print.go).
func NewBuildHandler() jsonrpc2.Handler {
	shared := new(handlerShared)
	return jsonrpc2.HandlerWithError((&BuildHandler{
		handlerShared: shared,
		lang:          &LangHandler{handlerShared: shared},
	}).handle)
}

// BuildHandler is a Go build server LSP/JSON-RPC handler that wraps a
// Go language server handler.
type BuildHandler struct {
	lang *LangHandler

	mu sync.Mutex
	handlerCommon
	*handlerShared
	init     *lspx.InitializeParams // set by "initialize" request
	depsDone bool                   // deps have been fetched and sent to the lang server
}

const (
	gopath     = "/"
	goroot     = "/goroot"
	gocompiler = "gc"

	// TODO(sqs): allow these to be customized. They're
	// fine for now, though.
	goos   = "linux"
	goarch = "amd64"
)

// reset clears all internal state in h.
func (h *BuildHandler) reset(init *lspx.InitializeParams, rootURI string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.handlerCommon.reset(rootURI); err != nil {
		return err
	}
	if err := h.handlerShared.reset(rootURI); err != nil {
		return err
	}
	h.init = init
	h.depsDone = false
	return nil
}

func (h *BuildHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	// Prevent any uncaught panics from taking the entire server down.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unexpected panic: %v", r)

			// Same as net/http
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("panic serving %v: %v\n%s", req.Method, r, buf)
			return
		}
	}()

	h.mu.Lock()
	if req.Method != "initialize" && h.init == nil {
		h.mu.Unlock()
		return nil, errors.New("server must be initialized")
	}
	h.mu.Unlock()
	if err := h.checkReady(); err != nil {
		if req.Method == "exit" {
			err = nil
		}
		return nil, err
	}

	h.initTracer(conn)
	span, ctx, err := h.spanForRequest(ctx, "build", req, opentracing.Tags{"mode": "go"})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
	}()

	switch {
	case req.Method == "initialize":
		if h.init != nil {
			return nil, errors.New("build server is already initialized")
		}
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lspx.InitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		// Determine the root import path of this workspace (e.g., "github.com/user/repo").
		if params.OriginalRootPath == "" {
			return nil, errors.New("unable to determine Go workspace root import path without due to empty root path")
		}
		u, err := uri.Parse(params.OriginalRootPath)
		if err != nil {
			return nil, err
		}
		var importPath string
		switch u.Scheme {
		case "git":
			importPath = path.Join(u.Host, strings.TrimSuffix(u.Path, ".git"), u.FilePath())
		default:
			return nil, fmt.Errorf("unrecognized originalRootPath: %q", u)
		}

		span.SetTag("originalRootPath", params.OriginalRootPath)

		// Sanity-check the import path.
		if importPath == "" || importPath != path.Clean(importPath) || strings.Contains(importPath, "..") || strings.HasPrefix(importPath, string(os.PathSeparator)) || strings.HasPrefix(importPath, "/") || strings.HasPrefix(importPath, ".") {
			return nil, fmt.Errorf("empty or suspicious import path: %q", importPath)
		}

		// Send "initialize" to the wrapped lang server.
		langInitParams := initializeParams{
			InitializeParams:     params.InitializeParams,
			NoOSFileSystemAccess: true,
			BuildContext: &initializeBuildContextParams{
				GOOS:       goos,
				GOARCH:     goarch,
				GOPATH:     gopath,
				GOROOT:     goroot,
				CgoEnabled: false,
				Compiler:   gocompiler,

				// TODO(sqs): We'd like to set this to true only for
				// the package we're analyzing (or for the whole
				// repo), but go/loader is insufficiently
				// configurable, so it applies it to the entire
				// program, which takes a lot longer and causes weird
				// error messages in the runtime package, etc. Disable
				// it for now.
				UseAllFiles: false,
			},
		}

		// Put all files in the workspace under a /src/IMPORTPATH
		// directory, such as /src/github.com/foo/bar, so that Go can
		// build it in GOPATH=/.
		rootFSPath := "/src/" + importPath
		langInitParams.RootPath = "file://" + rootFSPath
		if err := h.reset(&params, langInitParams.RootPath); err != nil {
			return nil, err
		}
		var langInitResp lsp.InitializeResult
		if err := h.callLangServer(ctx, conn, req.Method, req.Notif, langInitParams, &langInitResp); err != nil {
			return nil, err
		}
		return langInitResp, nil

	case req.Method == "shutdown":
		h.shutDown()
		return nil, nil

	case req.Method == "exit":
		conn.Close()
		return nil, nil

	default:
		// Pass the request onto the lang server.

		// Rewrite URI fields in params to refer to file paths inside
		// the GOPATH at the appropriate import path directory. E.g.:
		//
		//   file:///dir/file.go -> file:///src/github.com/user/repo/dir/file.go
		var urisInRequest []string // rewritten
		var params interface{}
		if req.Params != nil {
			if err := json.Unmarshal(*req.Params, &params); err != nil {
				return nil, err
			}
		}
		lspx.WalkURIFields(params, nil, func(uri string) string {
			if !strings.HasPrefix(uri, "file:///") {
				panic("URI in LSP request must be a file:/// URI, got " + uri)
			}
			path := strings.TrimPrefix(uri, "file://")
			path = pathpkg.Join(h.rootFSPath, path)
			if !pathHasPrefix(path, h.rootFSPath) {
				panic(fmt.Sprintf("file path %q must have prefix %q (file URI is %q, root URI is %q)", path, h.rootFSPath, uri, h.init.RootPath))
			}
			newURI := "file://" + path
			urisInRequest = append(urisInRequest, newURI) // collect
			return newURI
		})
		// Store back to req.Params to avoid 2 different versions of the data.
		if req.Params != nil {
			b, err := json.Marshal(params)
			if err != nil {
				return nil, err
			}
			req.Params = (*json.RawMessage)(&b)
		}

		// Immediately handle file system requests by adding them to
		// the VFS shared between the build and lang server.
		if isFileSystemRequest(req.Method) {
			if err := h.handleFileSystemRequest(ctx, req); err != nil {
				return nil, err
			}
			return nil, nil
		}

		// Fetch transitive dependencies for the named files, if this
		// is a language analysis request.
		fetchAndSendDeps := func() error {
			// TODO(sqs): this lock will have bad perf implications, make it finer
			h.mu.Lock()
			defer h.mu.Unlock()
			if h.depsDone {
				return nil
			}

			for _, uri := range urisInRequest {
				if err := h.fetchTransitiveDepsOfFile(ctx, uri); err != nil {
					log.Printf("Warning: fetching deps for Go file %q: %s.", uri, err)
				}
			}
			h.depsDone = true
			return nil
		}
		if err := fetchAndSendDeps(); err != nil {
			return nil, err
		}

		var result interface{}
		if err := h.callLangServer(ctx, conn, req.Method, req.Notif, params, &result); err != nil {
			return nil, err
		}

		// (Un-)rewrite URI fields in the result. E.g.:
		//
		//   file:///src/github.com/user/repo/dir/file.go -> file:///dir/file.go
		var walkErr error
		lspx.WalkURIFields(result, nil, func(uri string) string {
			newURI, err := h.rewriteURIFromLangServer(uri)
			if err != nil {
				walkErr = err
			}
			return newURI
		})
		if walkErr != nil {
			return nil, fmt.Errorf("%s (in Go language server response)", walkErr)
		}
		return result, nil
	}
}

func (h *BuildHandler) rewriteURIFromLangServer(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	if !u.IsAbs() {
		return "", fmt.Errorf("invalid relative URI %q", u)
	}

	switch u.Scheme {
	case "file":
		if !filepath.IsAbs(u.Path) {
			return "", fmt.Errorf("invalid relative file path in URI %q", uri)
		}

		// Refers to a file in the Go stdlib?
		if pathHasPrefix(u.Path, goroot) {
			fileInGoStdlib := pathTrimPrefix(u.Path, goroot)
			return "git://github.com/golang/go?" + runtime.Version() + "#" + fileInGoStdlib, nil
		}

		// Refers to a file in the same workspace?
		if pathHasPrefix(u.Path, h.rootFSPath) {
			pathInThisWorkspace := pathTrimPrefix(u.Path, h.rootFSPath)
			return "file:///" + pathInThisWorkspace, nil
		}

		// Refers to a file in the GOPATH (that's from another repo)?
		if gopathSrcDir := path.Join(gopath, "src"); pathHasPrefix(u.Path, gopathSrcDir) {
			p := pathTrimPrefix(u.Path, gopathSrcDir) // "github.com/foo/bar/baz/qux.go"

			// TODO(sqs) HACK to make
			// golang.org/x/... work. Better way is to record
			// where we fetched this from.
			if strings.HasPrefix(p, "golang.org/x/") {
				p = "github.com/golang/" + strings.TrimPrefix(p, "golang.org/x/")
			}
			if p == "google.golang.org/grpc" {
				p = "github.com/google/grpc-go"
			}

			// TODO(sqs): special-case github.com/ repos for now,
			// implement others soon...need to know where the
			// cutoff is between repo and subtree, which we
			// compute in deps.go.
			if strings.HasPrefix(p, "github.com/") {
				parts := strings.SplitN(p, "/", 4)
				if len(parts) >= 3 {
					var path string
					if len(parts) == 4 {
						path = parts[3]
					}
					return fmt.Sprintf("git://%s/%s/%s?HEAD#%s", parts[0], parts[1], parts[2], path), nil
				}
			}
		}

		return "unresolved:" + u.Path, nil
	default:
		return "", fmt.Errorf("invalid non-file URI %q", uri)
	}
}

// callLangServer sends the (usually modified) request to the wrapped
// Go language server. It
//
// Although bypasses the JSON-RPC wire protocol ( just sending it
// in-memory for simplicity/speed), it behaves in the same way as
// though the peer language server were remote. The conn is nil (and
// the request ID is zero'd out) to prevent the language server from
// breaking this abstraction.
func (h *BuildHandler) callLangServer(ctx context.Context, conn *jsonrpc2.Conn, method string, notif bool, params, result interface{}) error {
	req := jsonrpc2.Request{
		Method: method,
		Notif:  notif,
	}
	if err := req.SetParams(params); err != nil {
		return err
	}

	wrappedConn := &jsonrpc2ConnImpl{rewriteURI: h.rewriteURIFromLangServer, conn: conn}

	result0, err := h.lang.handle(ctx, wrappedConn, &req)
	if err != nil {
		return err
	}

	if !notif {
		// Don't pass the interface{} value, to avoid the build and
		// language servers from breaking the abstraction that they are in
		// separate memory spaces.
		b, err := json.Marshal(result0)
		if err != nil {
			return err
		}
		if result != nil {
			if err := json.Unmarshal(b, result); err != nil {
				return err
			}
		}
	}
	return nil
}

// jsonrpc2Conn is a limited interface to jsonrpc2.Conn. When the
// build server wraps the lang server, it provides this limited subset
// of methods. This interface exists to make it possible for the build
// server to provide the lang server with this limited connection
// handle.
type jsonrpc2Conn interface {
	Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) error
}

// jsonrpc2ConnImpl implements jsonrpc2Conn. See jsonrpc2Conn for more
// information.
type jsonrpc2ConnImpl struct {
	rewriteURI func(string) (string, error)
	conn       *jsonrpc2.Conn
}

func (c *jsonrpc2ConnImpl) Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) error {
	// Rewrite URIs from lang server (file:///src/github.com/foo/bar/f.go -> file:///f.go).
	switch method {
	case "textDocument/publishDiagnostics":
		params := params.(lsp.PublishDiagnosticsParams)

		newURI, err := c.rewriteURI(params.URI)
		if err != nil {
			return err
		}
		params.URI = newURI
		return c.conn.Notify(ctx, method, params, opt...)

	default:
		panic("build server wrapper for lang server notification sending does not support method " + method)
	}
}
