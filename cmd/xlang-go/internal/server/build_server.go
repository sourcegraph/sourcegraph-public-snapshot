package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	pathpkg "path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gregjones/httpcache"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/langserver"
	"github.com/sourcegraph/go-langserver/langserver/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	lsext "github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/pkg/gosrc"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

// Debug if true will cause extra logging information to be printed
var Debug = true

// NewHandler creates a new build server wrapping a (also newly
// created) Go language server. I.e., it creates a BuildHandler
// wrapping a LangHandler. The two handlers share a file system (in
// memory).
//
// The build server is responsible for things such as fetching
// dependencies, setting up the right file system structure and paths,
// and mapping local file system paths to logical URIs (e.g.,
// /goroot/src/fmt/print.go ->
// git://github.com/golang/go?go1.7.1#src/fmt/print.go).
func NewHandler() jsonrpc2.Handler {
	shared := &langserver.HandlerShared{Shared: true}
	h := &BuildHandler{
		HandlerShared: shared,
		lang: &langserver.LangHandler{
			HandlerShared: shared,
			DefaultConfig: langserver.Config{
				MaxParallelism: runtime.GOMAXPROCS(0),
			},
		},
	}
	shared.FindPackage = h.findPackageCached
	return jsonrpc2.HandlerWithError(h.handle)
}

// BuildHandler is a Go build server LSP/JSON-RPC handler that wraps a
// Go language server handler.
type BuildHandler struct {
	lang *langserver.LangHandler

	mu             sync.Mutex
	depURLMutex    *keyMutex
	gopathDeps     []*gosrc.Directory
	pinnedDepsOnce sync.Once
	pinnedDeps     pinnedPkgs
	findPkgMu      sync.Mutex // guards findPkg
	findPkg        map[findPkgKey]*findPkgValue
	langserver.HandlerCommon
	*langserver.HandlerShared
	init           *lspext.InitializeParams // set by "initialize" request
	rootImportPath string                   // root import path of the workspace (e.g., "github.com/foo/bar")
	cachingClient  *http.Client             // http.Client with a cache backed by the LSP Proxy, set by BuildHandler.reset()
}

// reset clears all internal state in h.
func (h *BuildHandler) reset(init *lspext.InitializeParams, conn *jsonrpc2.Conn, rootURI lsp.DocumentURI) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.findPkgMu.Lock()
	defer h.findPkgMu.Unlock()
	if err := h.HandlerCommon.Reset(rootURI); err != nil {
		return err
	}
	if err := h.HandlerShared.Reset(false); err != nil {
		return err
	}
	h.init = init
	h.cachingClient = &http.Client{Transport: httpcache.NewTransport(&lspCache{context.Background(), conn})}
	h.depURLMutex = newKeyMutex()
	h.gopathDeps = nil
	h.pinnedDepsOnce = sync.Once{}
	h.pinnedDeps = nil
	h.findPkg = nil
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
	if err := h.CheckReady(); err != nil {
		if req.Method == "exit" {
			err = nil
		}
		return nil, err
	}

	h.InitTracer(conn)
	span, ctx, err := h.SpanForRequest(ctx, "build", req, opentracing.Tags{"mode": "go"})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	if Debug && h.init != nil {
		var b []byte
		if req.Params != nil && !req.Notif {
			b = []byte(*req.Params)
		}
		log.Printf(">>> %s %s %s %s", h.init.OriginalRootURI, req.ID, req.Method, string(b))
		defer func(t time.Time) {
			log.Printf("<<< %s %s %s %dms", h.init.OriginalRootURI, req.ID, req.Method, time.Since(t).Nanoseconds()/int64(time.Millisecond))
		}(time.Now())
	}

	switch {
	case req.Method == "initialize":
		if h.init != nil {
			return nil, errors.New("build server is already initialized")
		}
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lspext.InitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		if Debug {
			var b []byte
			if req.Params != nil {
				b = []byte(*req.Params)
			}
			log.Printf(">>> %s %s %s %s", params.OriginalRootURI, req.ID, req.Method, string(b))
			defer func(t time.Time) {
				log.Printf("<<< %s %s %s %dms", params.OriginalRootURI, req.ID, req.Method, time.Since(t).Nanoseconds()/int64(time.Millisecond))
			}(time.Now())
		}

		// Determine the root import path of this workspace (e.g., "github.com/user/repo").
		span.SetTag("originalRootPath", params.OriginalRootURI)
		fs, err := remoteFS(ctx, conn, params.OriginalRootURI)
		if err != nil {
			return nil, err
		}

		langInitParams, err := determineEnvironment(ctx, fs, params)
		if err != nil {
			return nil, err
		}
		log.Printf("Detected root import path %q for %q", langInitParams.RootImportPath, params.OriginalRootURI)
		h.rootImportPath = langInitParams.RootImportPath
		if err := h.reset(&params, conn, langInitParams.Root()); err != nil {
			return nil, err
		}
		rootPath := strings.TrimPrefix(string(langInitParams.Root()), "file://")
		h.FS.Bind(rootPath, fs, "/", ctxvfs.BindAfter)
		var langInitResp lsp.InitializeResult
		if err := h.callLangServer(ctx, conn, req.Method, req.ID, langInitParams, &langInitResp); err != nil {
			return nil, err
		}
		return langInitResp, nil

	case req.Method == "shutdown":
		h.ShutDown()
		return nil, nil

	case req.Method == "exit":
		conn.Close()
		return nil, nil

	case req.Method == "$/cancelRequest":
		// Our caching layer is pretty bad, and can easily be poisened
		// if we cancel something. So we do not pass on cancellation
		// requests.
		return nil, nil

	case req.Method == "workspace/xpackages":
		return h.handleWorkspacePackages(ctx, conn, req)

	case req.Method == "workspace/xdependencies":
		// The same as h.fetchAndSendDepsOnce except it operates locally to the
		// request.
		fetchAndSendDepsOnces := make(map[string]*sync.Once) // key is file URI
		localFetchAndSendDepsOnce := func(fileURI string) *sync.Once {
			once, ok := fetchAndSendDepsOnces[fileURI]
			if !ok {
				once = new(sync.Once)
				fetchAndSendDepsOnces[fileURI] = once
			}
			return once
		}

		var (
			mu              sync.Mutex
			finalReferences []*lspext.DependencyReference
			references      = make(map[string]*lspext.DependencyReference)
		)
		emitRef := func(path string, r goDependencyReference) {
			// If the _reference_ to a definition is made from inside a
			// vendored package, or from outside of the repository itself,
			// exclude it.
			if util.IsVendorDir(path) || !util.PathHasPrefix(path, h.RootFSPath) {
				return
			}

			// If the package being referenced is defined in the repo, and
			// it is NOT a vendor package, then exclude it.
			if !r.vendor && util.PathHasPrefix(filepath.Join(gopath, "src", r.absolute), h.RootFSPath) {
				return
			}

			newURI, err := h.rewriteURIFromLangServer(lsp.DocumentURI("file://" + path))
			if err != nil {
				log.Printf("error rewriting URI from language server: %s", err)
				return
			}

			mu.Lock()
			defer mu.Unlock()

			existing, ok := references[r.absolute]
			if !ok {
				// Create a new dependency reference.
				ref := &lspext.DependencyReference{
					Attributes: r.attributes(),
					Hints: map[string]interface{}{
						"dirs": []string{string(newURI)},
					},
				}
				finalReferences = append(finalReferences, ref)
				references[r.absolute] = ref
				return
			}

			// Append to the existing dependency reference's dirs list.
			dirs := existing.Hints["dirs"].([]string)
			dirs = append(dirs, string(newURI))
			existing.Hints["dirs"] = dirs
		}

		// We need every transitive dependency, for every Go package in the
		// repository.
		var (
			w  = ctxvfs.Walk(ctx, h.RootFSPath, h.FS)
			dc = newDepCache()
		)
		dc.collectReferences = true
		for w.Step() {
			if path.Ext(w.Path()) == ".go" {
				d := path.Dir(w.Path())
				localFetchAndSendDepsOnce(d).Do(func() {
					if err := h.fetchTransitiveDepsOfFile(ctx, lsp.DocumentURI("file://"+d), dc); err != nil {
						log.Printf("Warning: fetching deps for dir %s: %s.", d, err)
					}
				})
			}
		}
		dc.references(emitRef, 1)
		return finalReferences, nil

	default:
		// Pass the request onto the lang server.

		// Rewrite URI fields in params to refer to file paths inside
		// the GOPATH at the appropriate import path directory. E.g.:
		//
		//   file:///dir/file.go -> file:///src/github.com/user/repo/dir/file.go
		var urisInRequest []lsp.DocumentURI // rewritten
		var params interface{}
		if req.Params != nil {
			if err := json.Unmarshal(*req.Params, &params); err != nil {
				return nil, err
			}
		}
		rewriteURIFromClient := func(uri lsp.DocumentURI) lsp.DocumentURI {
			if !strings.HasPrefix(string(uri), "file:///") {
				return uri // refers to a resource outside of this workspace
			}
			path := strings.TrimPrefix(string(uri), "file://")
			path = pathpkg.Join(h.RootFSPath, path)
			if !util.PathHasPrefix(path, h.RootFSPath) {
				panic(fmt.Sprintf("file path %q must have prefix %q (file URI is %q, root URI is %q)", path, h.RootFSPath, uri, h.init.RootPath))
			}
			newURI := lsp.DocumentURI("file://" + path)
			urisInRequest = append(urisInRequest, newURI) // collect
			return newURI
		}
		lspext.WalkURIFields(params, nil, rewriteURIFromClient)
		// Store back to req.Params to avoid 2 different versions of the data.
		if req.Params != nil {
			b, err := json.Marshal(params)
			if err != nil {
				return nil, err
			}
			req.Params = (*json.RawMessage)(&b)
		}

		// Immediately handle notifications. We do not have a response
		// to rewrite, so we can pass it on directly and avoid the
		// cost of marshalling again. NOTE: FS operations are frequent
		// and are notifications.
		if req.Notif {
			wrappedConn := &jsonrpc2ConnImpl{rewriteURI: h.rewriteURIFromLangServer, conn: conn}
			// Avoid extracting the tracer again, it is already attached to ctx.
			req.Meta = nil
			return h.lang.Handle(ctx, wrappedConn, req)
		}

		// workspace/symbol queries must have their `dir:` query filter
		// rewritten for github.com/golang/go due to its specialized directory
		// structure. e.g. `dir:src/net/http` should work, but the language
		// server will expect `dir:net/http` as any real/valid Go project will
		// have package paths align with the directory structure.
		if req.Method == "workspace/symbol" && strings.HasPrefix(string(h.init.OriginalRootURI), "git://github.com/golang/go") {
			var wsparams lsext.WorkspaceSymbolParams
			if err := json.Unmarshal(*req.Params, &wsparams); err != nil {
				return nil, err
			}
			q := langserver.ParseQuery(wsparams.Query)
			if q.Filter == langserver.FilterDir {
				// If the query does not start with `src/` and it is a request
				// for a stdlib dir, it should return no results (the filter is
				// dir, not package path).
				if gosrc.IsStdlibPkg(q.Dir) && !strings.HasPrefix(q.Dir, "src") {
					q.Dir = "sginvalid"
				} else {
					q.Dir = util.PathTrimPrefix(q.Dir, "src") // "src/net/http" -> "net/http"
				}
			}
			wsparams.Query = q.String()
			b, err := json.Marshal(wsparams)
			if err != nil {
				return nil, err
			}
			req.Params = (*json.RawMessage)(&b)
		}

		if req.Method == "workspace/xreferences" {
			// Parse the parameters and if a dirs hint is present, rewrite the
			// URIs.
			var p lsext.WorkspaceReferencesParams
			if err := json.Unmarshal(*req.Params, &p); err != nil {
				return nil, err
			}
			dirsHint, haveDirsHint := p.Hints["dirs"]
			if haveDirsHint {
				dirs := dirsHint.([]interface{})
				for i, dir := range dirs {
					dirs[i] = rewriteURIFromClient(lsp.DocumentURI(dir.(string)))
				}

				// Arbitrarily chosen limit on the number of directories that
				// may be searched by workspace/xreferences. Large repositories
				// like kubernetes would simply take too long (>15s) to fetch
				// their dependencies and typecheck them otherwise. This number
				// was chosen as a 'sweet-spot' based on kubernetes solely.
				if len(dirs) > 15 {
					dirs = dirs[:15]
				}
				dirsHint = dirs
				p.Hints["dirs"] = dirs
				b, err := json.Marshal(p)
				if err != nil {
					return nil, err
				}
				req.Params = (*json.RawMessage)(&b)
			}
		}

		var result interface{}
		if err := h.callLangServer(ctx, conn, req.Method, req.ID, req.Params, &result); err != nil {
			return nil, err
		}

		// (Un-)rewrite URI fields in the result. E.g.:
		//
		//   file:///src/github.com/user/repo/dir/file.go -> file:///dir/file.go
		var walkErr error
		lspext.WalkURIFields(result, nil, func(uri lsp.DocumentURI) lsp.DocumentURI {
			// HACK: Work around https://github.com/sourcegraph/sourcegraph/issues/10541 by
			// converting uri == "file://" (which is actually an empty URI in the langserver result)
			// to "file:///" instead of emitting an error. This will likely cause the result to be displayed
			// with an error on the client, but it's better than the whole
			// textDocument/implementation request failing.
			if req.Method == "textDocument/implementation" && (uri == "" || uri == "file://") {
				return "file:///"
			}

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

func (h *BuildHandler) rewriteURIFromLangServer(uri lsp.DocumentURI) (lsp.DocumentURI, error) {
	u, err := url.Parse(string(uri))
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
		if util.PathHasPrefix(u.Path, goroot) {
			fileInGoStdlib := util.PathTrimPrefix(u.Path, goroot)
			if h.rootImportPath == "" {
				// The workspace is the Go stdlib and this refers to
				// something in the Go stdlib, so let's use file:///
				// so that the LSP proxy adds our current rev, instead
				// of using runtime.Version() (which is not
				// necessarily the commit of the Go stdlib we're
				// analyzing).
				return lsp.DocumentURI("file:///" + fileInGoStdlib), nil
			}
			return lsp.DocumentURI("git://github.com/golang/go?" + gosrc.RuntimeVersion + "#" + fileInGoStdlib), nil
		}

		// Refers to a file in the same workspace?
		if util.PathHasPrefix(u.Path, h.RootFSPath) {
			pathInThisWorkspace := util.PathTrimPrefix(u.Path, h.RootFSPath)
			return lsp.DocumentURI("file:///" + pathInThisWorkspace), nil
		}

		// Refers to a file in the GOPATH (that's from another repo)?
		if gopathSrcDir := path.Join(gopath, "src"); util.PathHasPrefix(u.Path, gopathSrcDir) {
			p := util.PathTrimPrefix(u.Path, gopathSrcDir) // "github.com/foo/bar/baz/qux.go"

			// Go through the list of directories we have
			// mounted. We make a copy instead of holding the lock
			// in the for loop to avoid holding the lock for
			// longer than necessary.
			h.HandlerShared.Mu.Lock()
			deps := make([]*gosrc.Directory, len(h.gopathDeps))
			copy(deps, h.gopathDeps)
			h.HandlerShared.Mu.Unlock()
			var d *gosrc.Directory
			for _, dep := range deps {
				if strings.HasPrefix(p, dep.ProjectRoot) {
					d = dep
				}
			}
			if d != nil {
				rev := d.Rev
				if rev == "" {
					rev = "HEAD"
				}

				i := strings.Index(d.CloneURL, "://")
				if i >= 0 {
					repo := d.CloneURL[i+len("://"):]
					path := strings.TrimPrefix(strings.TrimPrefix(p, d.ProjectRoot), "/")

					// HACK
					// In some cases, we see import paths of the form "blah/blah.git" or "blah/blah.git/blah/blah".
					// The URI for the repository containing such a package is "blah/blah", so we strip the ".git"
					// from the location URI here. In addition, we strip any leading ".git/" from the path that
					// might get added as a side-effect of stripping the suffix.
					repo = strings.TrimSuffix(repo, ".git")
					path = strings.TrimPrefix(path, ".git/")

					return lsp.DocumentURI(fmt.Sprintf("%s://%s?%s#%s", d.VCS, repo, rev, path)), nil
				}
			}
		}

		return lsp.DocumentURI("unresolved:" + u.Path), nil
	default:
		return "", fmt.Errorf("invalid non-file URI %q", uri)
	}
}

// callLangServer sends the (usually modified) request to the wrapped Go
// language server. Do not send notifications via this interface! Rather just
// directly pass on the jsonrpc2.Request via h.lang.Handle.
//
// Although bypasses the JSON-RPC wire protocol ( just sending it
// in-memory for simplicity/speed), it behaves in the same way as
// though the peer language server were remote.
func (h *BuildHandler) callLangServer(ctx context.Context, conn *jsonrpc2.Conn, method string, id jsonrpc2.ID, params, result interface{}) error {
	req := jsonrpc2.Request{
		ID:     id,
		Method: method,
	}
	if err := req.SetParams(params); err != nil {
		return err
	}

	wrappedConn := &jsonrpc2ConnImpl{rewriteURI: h.rewriteURIFromLangServer, conn: conn}

	result0, err := h.lang.Handle(ctx, wrappedConn, &req)
	if err != nil {
		return err
	}

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
	return nil
}
