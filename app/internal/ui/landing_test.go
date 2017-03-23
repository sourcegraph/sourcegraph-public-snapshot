package ui

import (
	"testing"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

func TestCurlRepro(t *testing.T) {
	mode := "go"
	rootPath := "git://github.com/golang/go?master"
	method := "workspace/symbol"
	params := lsp.WorkspaceSymbolParams{
		Query: "dir:src/net/http Request",
		Limit: 100,
	}
	got := curlRepro(mode, rootPath, method, params)
	want := `Reproduce with: curl --data '[{"method":"initialize","params":{"rootPath":"git://github.com/golang/go?master","capabilities":{},"initializationOptions":{"mode":"go"}},"id":0,"jsonrpc":"2.0"},{"method":"workspace/symbol","params":{"query":"dir:src/net/http Request","limit":100},"id":1,"jsonrpc":"2.0"},{"method":"shutdown","id":2,"jsonrpc":"2.0"},{"method":"exit","id":0,"jsonrpc":"2.0"}]' https://sourcegraph.com/.api/xlang/workspace/symbol -i`
	if got != want {
		t.Log("got ", got)
		t.Error("want", want)
	}
}
