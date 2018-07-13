// Command cx-logdna is a Sourcegraph extension that decorates text documents with links to the
// relevant LogDNA view.
//
// It is very primitive now. It just generates LogDNA URLs when it sees a line containing
// "log15" and a string literal (which is used as the query term).
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
	logdnaOrgID = env.Get("LOGDNA_ORG_ID", "", "the LogDNA organization id (used for generating URLs)")
)

func main() {
	cxpmain.Main("cx-logdna", func() jsonrpc2.Handler { return jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError(handle)) })
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
		cap, err := cxp.ParseClientCapabilities(*req.Params)
		if err != nil {
			return nil, err
		}
		if !cap.Exec {
			return nil, errors.New("client does not support exec")
		}
		if cap.Decorations == nil || !cap.Decorations.Static {
			return nil, errors.New("client does not support decorations")
		}

		return cxp.InitializeResult{
			Capabilities: cxp.ServerCapabilities{
				DecorationsProvider: &cxp.DecorationsProviderServerCapabilities{DecorationsCapabilityOptions: cxp.DecorationsCapabilityOptions{Static: true}},
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

		fs := vfsutil.XRemoteFS{Conn: conn}
		f, err := fs.Open(ctx, "/"+path)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		loglines := findLogLines(data)
		decorations := make([]lspext.TextDocumentDecoration, len(loglines))
		for i, logline := range loglines {
			decorations[i] = lspext.TextDocumentDecoration{
				Range: lsp.Range{
					Start: lsp.Position{Line: logline.line},
					End:   lsp.Position{Line: logline.line},
				},
				After: &lspext.DecorationAttachmentRenderOptions{
					ContentText:     "Logs (LogDNA) » ",
					Color:           "rgba(255, 255, 255, 0.8)",
					BackgroundColor: "#db0a5b", // LogDNA brand color
					LinkURL: fmt.Sprintf("https://app.logdna.com/%s/logs/view?q=%s",
						logdnaOrgID,
						url.QueryEscape(logline.query),
					),
				},
			}
		}
		return decorations, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

type logline struct {
	line  int // 0-indexed
	query string
}

var loglinePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?:[Ll]og|[Ww]arn|[Ii]nfo|[Ee]rror|[Dd]ebug|[Pp]anic(?:f|ln)?|[Pp]rint(?:f|ln)?|[Ff]atal(?:f|ln)?)\("([^"]+)"`),
	regexp.MustCompile(`(?:[Ll]og|[Ww]arn|[Ii]nfo|[Ee]rror|[Dd]ebug|[Pp]anic(?:f|ln)?|[Pp]rint(?:f|ln)?|[Ff]atal(?:f|ln)?)\('([^']+)'`),
	regexp.MustCompile("(?:[Ll]og|[Ww]arn|[Ii]nfo|[Ee]rror|[Dd]ebug|[Pp]anic(?:f|ln)?|[Pp]rint(?:f|ln)?|[Ff]atal(?:f|ln)?)\\(`([^`]+)`"),
}

func findLogLines(text []byte) []logline {
	var found []logline
	for i, line := range bytes.Split(text, []byte("\n")) {
		for _, loglinePattern := range loglinePatterns {
			for _, m := range loglinePattern.FindAllSubmatch(line, -1) {
				found = append(found, logline{line: i, query: string(m[1])})
			}
		}
	}
	return found
}
