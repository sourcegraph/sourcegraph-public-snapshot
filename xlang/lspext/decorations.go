package lspext

import "github.com/sourcegraph/go-langserver/pkg/lsp"

// TextDocumentDecoration represents a decoration to apply to a text document.
type TextDocumentDecoration struct {
	Range           lsp.Range `json:"range"`
	IsWholeLine     bool      `json:"isWholeLine,omitempty"`
	BackgroundColor string    `json:"backgroundColor,omitempty"`
}
