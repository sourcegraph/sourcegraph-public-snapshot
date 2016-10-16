package xlang

import (
	"testing"

	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"
)

func TestGuessTrackedErrorURL(t *testing.T) {
	got, err := guessTrackedErrorURL(trackedError{
		Error:  "jsonrpc2: code 0 message: type/object not found at {Line:4 Character:9}",
		Method: "textDocument/hover",
		Mode:   "go",
		Params: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: "git://github.com/gorilla/mux?757bef944d0f21880861c2dd9c871ca543023cba#mux.go",
			},
			Position: lsp.Position{Line: 4, Character: 9},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "github.com/gorilla/mux@757bef944d0f21880861c2dd9c871ca543023cba/-/blob/mux.go#L5:10"
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
