package langp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2/jsonrpc2test"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func init() {
	InitMetrics("test")
}

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
		PrepareRepo: func(update bool, workspace, repo, commit string) error { return nil },
		PrepareDeps: func(update bool, workspace, repo, commit string) error { return nil },
		ResolveFile: func(workspace, repo, commit, uri string) (*File, error) {
			if !strings.HasPrefix(uri, "file:///") {
				return nil, fmt.Errorf("uri does not start with file:/// : %s", uri)
			}
			path := uri[8:]
			deps := map[string]string{
				"github.com/gorilla/mux": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				"github.com/golang/go":   "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			}
			deps[repo] = commit
			repo = "unknown"
			commit = "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"
			for r, c := range deps {
				if strings.HasPrefix(path, r) {
					repo = r
					commit = c
					path = path[len(r)+1:]
				}
			}
			return &File{Repo: repo, Commit: commit, Path: path}, nil
		},
		FileURI: func(repo, commit, file string) string { return "file:///" + filepath.Join(repo, file) },
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
				TextDocument: lsp.TextDocumentIdentifier{URI: "file:///github.com/foo/bar/baz.go"},
				Position: lsp.Position{
					Line:      12,
					Character: 34,
				},
			},

			LSPResponseResult: []lsp.Location{{
				URI: "file:///github.com/foo/bar/baz.go",
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
				TextDocument: lsp.TextDocumentIdentifier{URI: "file:///github.com/foo/bar/baz.go"},
				Position: lsp.Position{
					Line:      12,
					Character: 34,
				},
			},

			LSPResponseResult: lsp.Hover{
				Contents: []lsp.MarkedString{{
					Language: "go",
					Value:    "NewRouter func() *Router",
				}, {
					Language: "text/html",
					Value:    "\u003cp\u003e\nNewRouter returns a new router instance.\n\u003c/p\u003e\n",
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
				Title:   "NewRouter func() *Router",
				DocHTML: "\u003cp\u003e\nNewRouter returns a new router instance.\n\u003c/p\u003e\n",
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
			WantResponse: &ExternalRefs{Defs: []*DefSpec{{
				Repo:     "github.com/gorilla/mux",
				Commit:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				UnitType: "GoPackage",
				Unit:     "github.com/gorilla/mux",
				Path:     "NewRouter",
			}, {
				Repo:     "github.com/golang/go",
				Commit:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
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
				Name: "Baz",
				Kind: 12,
				Location: lsp.Location{
					URI: "file:///github.com/foo/bar/baz.go",
					// Range ignored
				},
				ContainerName: "github.com/foo/bar",
			}, {
				Name: "New",
				Kind: 12,
				Location: lsp.Location{
					URI: "file:///github.com/foo/bar/baz.go",
					// Range ignored
				},
				ContainerName: "github.com/foo/bar",
			}},
			WantResponse: &ExportedSymbols{Symbols: []*Symbol{{
				DefSpec: DefSpec{
					Repo:     "github.com/foo/bar",
					Commit:   "deadbeef",
					UnitType: "GoPackage",
					Unit:     "github.com/foo/bar",
					Path:     "Baz",
				},
				Name: "Baz",
				Kind: "func",
				File: "baz.go",
			}, {
				DefSpec: DefSpec{
					Repo:     "github.com/foo/bar",
					Commit:   "deadbeef",
					UnitType: "GoPackage",
					Unit:     "github.com/foo/bar",
					Path:     "New",
				},
				Name: "New",
				Kind: "func",
				File: "baz.go",
			}}},
			Got: &ExportedSymbols{},
		},

		// ServeLocalRefs
		{
			Path: "/local-refs",
			Request: &Position{
				Repo:      "github.com/foo/bar",
				Commit:    "deadbeef",
				File:      "baz.go",
				Line:      12,
				Character: 34,
			},
			WantLSPMethod: "textDocument/references",
			WantLSPParam: &lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: "file:///github.com/foo/bar/baz.go"},
				Position: lsp.Position{
					Line:      12,
					Character: 34,
				},
			},

			LSPResponseResult: []lsp.Location{{
				URI: "file:///github.com/foo/bar/baz.go",
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
			}, {
				URI: "file:///github.com/foo/bar/baz.go",
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      4,
						Character: 2,
					},
					End: lsp.Position{
						Line:      4,
						Character: 5,
					},
				},
			}},
			WantResponse: &LocalRefs{Refs: []*Range{
				{
					Repo:           "github.com/foo/bar",
					Commit:         "deadbeef",
					File:           "baz.go",
					StartLine:      1,
					StartCharacter: 2,
					EndLine:        1,
					EndCharacter:   5,
				}, {
					Repo:           "github.com/foo/bar",
					Commit:         "deadbeef",
					File:           "baz.go",
					StartLine:      4,
					StartCharacter: 2,
					EndLine:        4,
					EndCharacter:   5,
				},
			}},
			Got: &LocalRefs{},
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
			t.Fatalf("got\n%s, want\n%s", marshal(t, c.Got), marshal(t, c.WantResponse))
		}
	}
}

func marshal(t *testing.T, v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
