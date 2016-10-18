package ctags

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/ctags/parser"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

func vslog(out ...string) {
	os.Stderr.WriteString(strings.Join(out, "\n") + "\n")
}

var emptyArray = make([]string, 0)

type InitParams struct {
	RootPath         string
	OriginalRootPath string
}

// Handler represents an LSP handler for one user.
type Handler struct {
	// fs is the virtual filesystem backed by xlang infrastrucuture.
	fs ctxvfs.FileSystem

	// tagsMu protects tags. We want to be careful to not run ctags more than
	// once for one project, so this is used in the get tags method.
	tagsMu sync.Mutex

	// tags is the Go form of the ctags output for this project. We compute and
	// save it so that we don't have to parse the ctags file each time, and so
	// we don't have to store as much state on disk.
	tags []parser.Tag
}

var ErrMustInit = errors.New("initialize must be called before other methods")

func (h *Handler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (_ interface{}, err error) {
	// Prevent any uncaught panics from taking the entire server down.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unexpected panic: %v", r)

			// Same as net/http
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			vslog(string(buf))
			log.Printf("panic serving %v: %v\n%s", req.Method, r, buf)
			return
		}
	}()

	operationName := "LS Serve: " + req.Method
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	defer span.Finish()

	h.fs = &vfsutil.RemoteProxyFS{Conn: conn}

	switch req.Method {
	case "initialize":
		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				WorkspaceSymbolProvider: true,
			},
		}, nil

	case "shutdown":
		return nil, nil

	case "workspace/symbol":
		var params lsp.WorkspaceSymbolParams
		if err = json.Unmarshal(*req.Params, &params); err != nil {
			vslog(err.Error())
			return
		}
		s, err := h.handleSymbol(ctx, req, params)
		if err != nil {
			vslog(err.Error())
			return nil, err
		}
		if s == nil {
			return emptyArray, nil
		}
		return s, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}
