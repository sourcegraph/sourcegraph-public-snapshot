package lspext

import "github.com/sourcegraph/go-langserver/pkg/lsp"

// See https://github.com/sourcegraph/language-server-protocol/pull/4.

// ContentParams is the input for 'textDocument/content'. The response is a
// 'TextDocumentItem'.
type ContentParams struct {
	TextDocument lsp.TextDocumentIdentifier `json:"textDocument"`
}

// FilesParams is the input for 'workspace/xfiles'. The response is '[]TextDocumentIdentifier'
type FilesParams struct {
	Base string `json:"patterns,omitempty"`
}
