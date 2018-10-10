package xlang

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/sourcegraph/go-lsp"
)

func TestRemoteOneShotClientRequest(t *testing.T) {
	var (
		wantRequestBody = `[{"method":"initialize","params":{"rootUri":"git://github.com/gorilla/mux?4dbd923b0c9e99ff63ad54b0e9705ff92d3cdb06","capabilities":{"workspace":{},"textDocument":{"completion":{"completionItemKind":{},"completionItem":{}}}},"initializationOptions":{"mode":"go"},"mode":"go"},"id":0,"jsonrpc":"2.0"},{"method":"textDocument/hover","params":{"textDocument":{"uri":"git://github.com/gorilla/mux?4dbd923b0c9e99ff63ad54b0e9705ff92d3cdb06#mux.go"},"position":{"line":93,"character":10}},"id":1,"jsonrpc":"2.0"},{"method":"shutdown","id":2,"jsonrpc":"2.0"},{"method":"exit","jsonrpc":"2.0"}]`
		gotRequestBody  string

		wantPath = "/.api/xlang/textDocument/hover"
		gotPath  string

		responseBody = `[{"id":0,"result":{"capabilities":{"textDocumentSync":2,"hoverProvider":true,"signatureHelpProvider":{"triggerCharacters":["(",","]},"definitionProvider":true,"referencesProvider":true,"documentSymbolProvider":true,"workspaceSymbolProvider":true,"implementationProvider":true,"documentFormattingProvider":true,"xworkspaceReferencesProvider":true,"xdefinitionProvider":true,"xworkspaceSymbolByProperties":true}},"jsonrpc":"2.0"},{"id":1,"result":{"contents":[{"language":"go","value":"struct field MatchErr error"},"MatchErr is set to appropriate matching error It is set to ErrMethodMismatch if there is a mismatch in the request method and route method \n\n"],"range":{"end":{"character":18,"line":93},"start":{"character":10,"line":93}}},"jsonrpc":"2.0"},{"id":2,"result":null,"jsonrpc":"2.0"},null]`
	)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		gotRequestBody = string(b)
		gotPath = r.URL.Path
		fmt.Fprintf(w, responseBody)
	}))
	defer ts.Close()

	ctx := context.Background()
	remote, _ := url.Parse(ts.URL)
	mode := "go"
	rootURI := lsp.DocumentURI("git://github.com/gorilla/mux?4dbd923b0c9e99ff63ad54b0e9705ff92d3cdb06")
	method := "textDocument/hover"
	params := &lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "git://github.com/gorilla/mux?4dbd923b0c9e99ff63ad54b0e9705ff92d3cdb06#mux.go"},
		Position:     lsp.Position{Line: 93, Character: 10},
	}
	var result lsp.Hover
	err := RemoteOneShotClientRequest(ctx, remote, mode, rootURI, method, params, &result)
	if err != nil {
		t.Fatal(err)
	}

	if gotRequestBody != wantRequestBody {
		t.Errorf("Unexpected request body:\ngot:  %s\nwant: %s", gotRequestBody, wantRequestBody)
	}
	if gotPath != wantPath {
		t.Errorf("Unexpected path: got %q != %q", gotPath, wantPath)
	}

	wantHover := lsp.Hover{
		Contents: []lsp.MarkedString{
			{
				Language: "go",
				Value:    "struct field MatchErr error",
			},
			lsp.RawMarkedString("MatchErr is set to appropriate matching error It is set to ErrMethodMismatch if there is a mismatch in the request method and route method \n\n"),
		},
		Range: &lsp.Range{
			Start: lsp.Position{
				Line:      93,
				Character: 10,
			},
			End: lsp.Position{
				Line:      93,
				Character: 18,
			},
		},
	}
	if !reflect.DeepEqual(result, wantHover) {
		t.Errorf("unexpected result:\ngot:  %+v\nwant: %+v", result, wantHover)
	}
}
