package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
)

func serveJumpToDef(w http.ResponseWriter, r *http.Request) error {
	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}

	file := r.URL.Query().Get("file")

	line, err := strconv.Atoi(r.URL.Query().Get("line"))
	if err != nil {
		return err
	}

	character, err := strconv.Atoi(r.URL.Query().Get("character"))
	if err != nil {
		return err
	}

	var response lsp.Location

	// TODO deprecate this endpoint. For now we send a request to our
	// xlang endpoint.
	var locations []lsp.Location
	err = textDocumentPositionRequest{
		Method:    "textDocument/definition",
		Repo:      repo.URI,
		Commit:    repoRev.CommitID,
		File:      file,
		Line:      line,
		Character: character,
	}.Serve(r.Context(), &locations)
	if err == nil && len(locations) >= 1 {
		response = locations[0]
	}
	w.Header().Set("cache-control", "private, max-age=60")
	return writeJSON(w, response)
}

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
			ID:     0,
			Method: "initialize",
			Params: mustMarshal(&xlang.ClientProxyInitializeParams{
				InitializeParams: lsp.InitializeParams{
					RootPath: fmt.Sprintf("git://%s?%s", o.Repo, o.Commit),
				},
				Mode: "go",
			}),
		},
		{
			ID:     1,
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
			ID:     2,
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
