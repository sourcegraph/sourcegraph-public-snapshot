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

		// ServeHover
		{
			Path: "/hover",
			Request: &Position{
				Repo:      "github.com/foo/bar",
				Commit:    "deadbeef",
				File:      "baz.go",
				Line:      12,
				Character: 34,
			},
			WantLSPMethod: "textDocument/hover",
			WantLSPParam: &lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: "github.com/foo/bar/baz.go"},
				Position: lsp.Position{
					Line:      12,
					Character: 34,
				},
			},

			LSPResponseResult: lsp.Hover{
				Contents: []lsp.MarkedString{{
					Language: "go",
					Value:    "NewRouter func() *Router",
				}},
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
			},
			WantResponse: &Hover{
				Contents: []HoverContent{{
					Type:  "go",
					Value: "NewRouter func() *Router",
				}},
			},
			Got: &Hover{},
		},

		// ServeExternalRefs
		{
			Path: "/external-refs",
			Request: &RepoRev{
				Repo:   "github.com/foo/bar",
				Commit: "deadbeef",
			},
			WantLSPMethod: "workspace/symbol",
			WantLSPParam: &lsp.WorkspaceSymbolParams{
				Query: "external github.com/foo/bar/...",
			},

			LSPResponseResult: []lsp.SymbolInformation{{
				Name:          "NewRouter",
				Kind:          12,
				Location:      lsp.Location{}, // Ignored
				ContainerName: "github.com/gorilla/mux",
			}, {
				Name:          "Printf",
				Kind:          12,
				Location:      lsp.Location{}, // Ignored
				ContainerName: "fmt",
			}},
			WantResponse: &ExternalRefs{Defs: []DefSpec{{
				Repo:     "github.com/gorilla/mux",
				Commit:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // TODO to translate deps commit
				UnitType: "GoPackage",
				Unit:     "github.com/gorilla/mux",
				Path:     "NewRouter",
			}, {
				Repo:     "github.com/golang/go",
				Commit:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // TODO to translate deps commit
				UnitType: "GoPackage",
				Unit:     "fmt",
				Path:     "Printf",
			}}},
			Got: &ExternalRefs{},
		},

		// ServeExportedSymbols
		{
			Path: "/exported-symbols",
			Request: &RepoRev{
				Repo:   "github.com/foo/bar",
				Commit: "deadbeef",
			},
			WantLSPMethod: "workspace/symbol",
			WantLSPParam: &lsp.WorkspaceSymbolParams{
				Query: "exported github.com/foo/bar/...",
			},

			LSPResponseResult: []lsp.SymbolInformation{{
				Name:          "Baz",
				Kind:          12,
				Location:      lsp.Location{}, // Ignored
				ContainerName: "github.com/foo/bar",
			}, {
				Name:          "New",
				Kind:          12,
				Location:      lsp.Location{}, // Ignored
				ContainerName: "github.com/foo/bar",
			}},
			WantResponse: &ExportedSymbols{Defs: []DefSpec{{
				Repo:     "github.com/foo/bar",
				Commit:   "deadbeef",
				UnitType: "GoPackage",
				Unit:     "github.com/foo/bar",
				Path:     "Baz",
			}, {
				Repo:     "github.com/foo/bar",
				Commit:   "deadbeef",
				UnitType: "GoPackage",
				Unit:     "github.com/foo/bar",
				Path:     "New",
			}}},
			Got: &ExportedSymbols{},
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
		resp, err := http.Post(ts.URL+c.Path, "application/json", bytes.NewReader(requestBody))
		if err != nil {
			t.Fatal(err)
		}
		err = json.NewDecoder(resp.Body).Decode(c.Got)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(c.Got, c.WantResponse) {
			t.Fatalf("got\n%+#v, want\n%+#v", c.Got, c.WantResponse)
		}
	}
}
