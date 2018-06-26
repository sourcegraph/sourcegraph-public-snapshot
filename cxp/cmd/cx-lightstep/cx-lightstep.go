// Command cx-lightstep is a Sourcegraph extension that decorates text documents with links to the
// relevant LightStep trace view.
//
// It is very primitive now. It just generates LightStep search URLs when it sees a line containing
// "StartSpan" and a string literal (which is used as the query term).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/cxp"
	"github.com/sourcegraph/sourcegraph/cxp/pkg/cxpmain"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

//docker:user sourcegraph

var (
	lightstepProject = env.Get("LIGHTSTEP_PROJECT", "", "the LightStep project name (used for generating URLs)")
)

func main() {
	cxpmain.Main("cx-lightstep", func() jsonrpc2.Handler { return jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError(handle)) })
}

func handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CXP: "+req.Method)
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
		cap, err := cxp.ParseExperimentalClientCapabilities(*req.Params)
		if err != nil {
			return nil, err
		}
		if !cap.Exec {
			return nil, errors.New("client does not support exec")
		}
		if !cap.Decorations {
			return nil, errors.New("client does not support decorations")
		}

		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				Experimental: lspext.ExperimentalServerCapabilities{
					DecorationsProvider: true,
				},
			},
		}, nil

	case "shutdown", "exit":
		return nil, nil

	case "textDocument/decorations":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lspext.TextDocumentDecorationsParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		uri, err := url.Parse(string(params.TextDocument.URI))
		if err != nil {
			return nil, err
		}
		path := strings.TrimPrefix(uri.Path, "/")

		fs := vfsutil.RemoteFS(conn)
		f, err := fs.Open(ctx, "/"+path)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		spans := findSpans(data)
		decorations := make([]lspext.TextDocumentDecoration, len(spans))
		for i, span := range spans {
			decorations[i] = lspext.TextDocumentDecoration{
				Range: lsp.Range{
					Start: lsp.Position{Line: span.line},
					End:   lsp.Position{Line: span.line},
				},
				After: &lspext.DecorationAttachmentRenderOptions{
					ContentText:     "Live traces (LightStep) » ",
					Color:           "rgba(255, 255, 255, 0.8)",
					BackgroundColor: "#2925ff", // LightStep brand color
					LinkURL: fmt.Sprintf("https://app.lightstep.com/%s/live?q=%s",
						lightstepProject,
						url.QueryEscape(fmt.Sprintf("operation:%q", span.query)),
					),
				},
			}
		}
		return decorations, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

type span struct {
	line  int // 0-indexed
	query string
}

var spanPattern = regexp.MustCompile(`StartSpan[^"]+"([^"]+)"`)

func findSpans(text []byte) []span {
	var found []span
	for i, line := range bytes.Split(text, []byte("\n")) {
		for _, m := range spanPattern.FindAllSubmatch(line, -1) {
			found = append(found, span{line: i, query: string(m[1])})
		}
	}
	return found
}
