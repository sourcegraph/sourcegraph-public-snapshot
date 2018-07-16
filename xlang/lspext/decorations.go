package lspext

import "github.com/sourcegraph/go-langserver/pkg/lsp"

// TextDocumentDecorationParams contains the parameters to the textDocument/decorations method.
type TextDocumentDecorationParams struct {
	TextDocument lsp.TextDocumentIdentifier `json:"textDocument"`
}

// TextDocumentDecoration represents a decoration to apply to a text document.
type TextDocumentDecoration struct {
	After           *DecorationAttachmentRenderOptions `json:"after,omitempty"`
	BackgroundColor string                             `json:"backgroundColor,omitempty"`
	Border          string                             `json:"border,omitempty"`
	BorderColor     string                             `json:"borderColor,omitempty"`
	BorderWidth     string                             `json:"borderWidth,omitempty"`
	IsWholeLine     bool                               `json:"isWholeLine,omitempty"`
	Range           lsp.Range                          `json:"range"`
}

// DecorationAttachmentRenderOptions defines a decoration attachment in a TextDocumentDecoration.
type DecorationAttachmentRenderOptions struct {
	BackgroundColor string `json:"backgroundColor,omitempty"`
	Color           string `json:"color,omitempty"`
	ContentText     string `json:"contentText,omitempty"`
	HoverMessage    string `json:"hoverMessage,omitempty"`
	LinkURL         string `json:"linkURL,omitempty"`
}
