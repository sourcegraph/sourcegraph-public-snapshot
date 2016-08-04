package langp

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2/jsonrpc2test"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func TestTranslator(t *testing.T) {
	workDir, err := ioutil.TempDir("", "TestServe")
	if err != nil {
		t.Fatal("TempDir failed: ", err)
	}
	defer os.RemoveAll(workDir)

	lspMock := jsonrpc2test.NewServer()
	lspMock.T = t
	defer lspMock.Close()

	ts := httptest.NewServer(New(&Translator{
		Addr:        lspMock.Addr,
		WorkDir:     workDir,
		PrepareRepo: func(workspace, repo, commit string) error { return nil },
		PrepareDeps: func(workspace, repo, commit string) error { return nil },
		FileURI:     func(repo, commit, file string) string { return filepath.Join(repo, file) },
	}))
	defer ts.Close()

	jsonRawMessage := func(v interface{}) *json.RawMessage {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		r := json.RawMessage(b)
		return &r
	}

	cases := []struct {
		Path          string
		Request       interface{}
		WantLSPMethod string
		WantLSPParam  interface{}

		LSPResponseResult interface{}
		WantResponse      interface{}
		Got               interface{}
	}{
		// ServeDefinition
		{
			Path: "/definition",
			Request: &Position{
				Repo:      "github.com/foo/bar",
				Commit:    "deadbeef",
				File:      "baz.go",
				Line:      12,
				Character: 34,
			},
			WantLSPMethod: "textDocument/definition",
			WantLSPParam: &lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: "github.com/foo/bar/baz.go"},
				Position: lsp.Position{
					Line:      12,
					Character: 34,
				},
			},

			LSPResponseResult: []lsp.Location{{
				URI: "baz.go",
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      1,
						Character: 2,
					},
					End: lsp.Position{
						Line:      1,
						Character: 5,
					},
				},
			}},
			WantResponse: &Range{
				Repo:           "github.com/foo/bar",
				Commit:         "deadbeef",
				File:           "baz.go",
				StartLine:      1,
				StartCharacter: 2,
				EndLine:        1,
				EndCharacter:   5,
			},
			Got: &Range{},
		},
	}

	for _, c := range cases {
		lspMock.WantRequest["1"] = &jsonrpc2.Request{
			ID:           "1",
			Method:       c.WantLSPMethod,
			Params:       jsonRawMessage(c.WantLSPParam),
			Notification: false,
			JSONRPC:      "2.0",
		}
		lspMock.Response["1"] = &jsonrpc2.Response{
			ID:      "1",
			Result:  jsonRawMessage(c.LSPResponseResult),
			JSONRPC: "2.0",
		}

		requestBody, err := json.Marshal(c.Request)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := http.Post(ts.URL+"/definition", "application/json", bytes.NewReader(requestBody))
		if err != nil {
			t.Fatal(err)
		}
		err = json.NewDecoder(resp.Body).Decode(c.Got)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(c.Got, c.WantResponse) {
			t.Fatalf("got %+#v, want %+#v", c.Got, c.WantResponse)
		}
	}
}
