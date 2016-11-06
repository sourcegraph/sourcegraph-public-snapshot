package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
)

type textDocumentPositionRequest struct {
	Method string

	Repo   string
	Commit string
	File   string
	// Line is 0 based
	Line int
	// Character is 0 based
	Character int
}

func (o textDocumentPositionRequest) Serve(ctx context.Context, result interface{}) error {
	if !strings.HasSuffix(o.File, ".go") {
		return &errcode.HTTPErr{
			Status: http.StatusNotFound,
			Err:    errors.New("currently only go is supported"),
		}
	}

	mustMarshal := func(v interface{}) *json.RawMessage {
		b, _ := json.Marshal(v)
		m := json.RawMessage(b)
		return &m
	}
	reqs := []jsonrpc2.Request{
		{
			ID:     jsonrpc2.ID{Num: 0},
			Method: "initialize",
			Params: mustMarshal(&xlang.ClientProxyInitializeParams{
				InitializeParams: lsp.InitializeParams{
					RootPath: fmt.Sprintf("git://%s?%s", o.Repo, o.Commit),
				},
				Mode: "go",
			}),
		},
		{
			ID:     jsonrpc2.ID{Num: 1},
			Method: o.Method,
			Params: mustMarshal(&lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: fmt.Sprintf("git://%s?%s#%s", o.Repo, o.Commit, o.File),
				},
				Position: lsp.Position{
					Line:      o.Line,
					Character: o.Character,
				},
			}),
		},
		{
			ID:     jsonrpc2.ID{Num: 2},
			Method: "shutdown",
		},
		{
			Method: "exit",
			Notif:  true,
		},
	}
	body, err := json.Marshal(reqs)
	if err != nil {
		return err
	}
	w := httptest.NewRecorder()
	err = serveXLangMethod(ctx, w, o.Method, bytes.NewReader(body))
	if err != nil {
		return err
	}

	resps := make([]*jsonrpc2.Response, len(reqs))
	err = json.Unmarshal(w.Body.Bytes(), &resps)
	if err != nil {
		return err
	}
	if len(resps) < 2 {
		return errors.New("not enough responses")
	}
	if resps[1].Result == nil {
		return errors.New("nil result")
	}
	return json.Unmarshal([]byte(*resps[1].Result), result)
}
