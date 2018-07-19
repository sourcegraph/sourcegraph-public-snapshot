// Command cx-line-age is a Sourcegraph extension that decorates text documents based on the age of
// each line (since its last commit).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/google/pprof/profile"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/cxp"
	"github.com/sourcegraph/sourcegraph/cxp/pkg/cxpmain"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/uri"
)

//docker:user sourcegraph

// TODO support multiple repos/projects/environments/apps
// filtering rules based on dir, regex, initialize params?

func main() {
	cxpmain.Main("cx-line-age", func() jsonrpc2.Handler { return jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError((&handler{}).handle)) })
}

type handler struct {
	mu       sync.Mutex
	rootURI  *uri.URI
	settings extensionSettings
}

type extensionSettings struct {
	PprofEndpoint string `json:"pprofEndpoint,omitempty"`
}

func (h *handler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "CXP: "+req.Method)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())

		}
		span.Finish()
	}()

	switch req.Method {
	case "initialize":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		cap, err := cxp.ParseClientCapabilities(*req.Params)
		if err != nil {
			return nil, err
		}
		if cap.Decoration == nil || !cap.Decoration.Static {
			return nil, errors.New("client does not support decorations")
		}

		var params cxp.InitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		var settings extensionSettings
		if merged := params.InitializationOptions.Settings.Merged; merged != nil {
			if err := json.Unmarshal(*merged, &settings); err != nil {
				return nil, err
			}
		}
		h.mu.Lock()
		h.settings = settings
		h.mu.Unlock()

		rootURI, err := uri.Parse(string(params.OriginalRootURI))
		if err != nil {
			return nil, err
		}
		h.mu.Lock()
		h.rootURI = rootURI
		h.mu.Unlock()

		return cxp.InitializeResult{
			Capabilities: cxp.ServerCapabilities{
				DecorationProvider: &cxp.DecorationProviderServerCapabilities{DecorationCapabilityOptions: cxp.DecorationCapabilityOptions{Static: true}},
			},
		}, nil

	case "shutdown", "exit":
		return nil, nil

	case "textDocument/hover":
		return nil, nil

	case "textDocument/decoration":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lspext.TextDocumentDecorationParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		docURI, err := url.Parse(string(params.TextDocument.URI))
		if err != nil {
			return nil, err
		}
		docPath := strings.TrimPrefix(docURI.Path, "/")

		pprofProfileFile, err := ioutil.TempFile(os.TempDir(), "")
		if err != nil {
			return nil, errors.Wrap(err, "error creating a temporary file for the pprof profile")
		}
		defer os.Remove(pprofProfileFile.Name())

		if h.settings.PprofEndpoint == "" {
			return nil, errors.New("Must specify pprofEndpoint")
		}

		err = exec.Command("pprof", "-proto", "-output", pprofProfileFile.Name(), h.settings.PprofEndpoint+"/goroutine").Run()
		if err != nil {
			return nil, errors.Wrap(err, "error running pprof")
		}

		pprofProfile, err := profile.Parse(pprofProfileFile)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing pprof profile")
		}

		type location struct {
			Filename string
			Line     int64
		}

		hitsByLocation := make(map[location]int64)
		for _, sample := range pprofProfile.Sample {
			for _, pprofLocation := range sample.Location {
				loc := location{
					Filename: pprofLocation.Line[0].Function.Filename,
					Line:     pprofLocation.Line[0].Line,
				}
				if _, ok := hitsByLocation[loc]; ok {
					hitsByLocation[loc] += sample.Value[0]
				} else {
					hitsByLocation[loc] = sample.Value[0]
				}
			}
		}

		lineDecorations := []lspext.TextDocumentDecoration{}
		for loc := range hitsByLocation {
			if strings.HasSuffix(loc.Filename, h.rootURI.Path+"/"+docPath) {
				blueHue := 240
				lineDecorations = append(lineDecorations, lspext.TextDocumentDecoration{
					BackgroundColor: fmt.Sprintf("hsla(%d, 100%%, 50%%, 0.1)", blueHue),
					Range: lsp.Range{
						Start: lsp.Position{Line: int(loc.Line - 1)},
						End:   lsp.Position{Line: int(loc.Line - 1)},
					},
					IsWholeLine: true,
				})
			}
		}

		return lineDecorations, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}
