package lspext

import "github.com/sourcegraph/go-langserver/pkg/lsp"

// PartialResultParams is the input for "$/partialResult", a notification.
type PartialResultParams struct {
	// ID is the jsonrpc2 ID of the request we are returning partial
	// results for.
	ID lsp.ID `json:"id"`

	// Patch is a JSON patch as specified at http://jsonpatch.com/
	//
	// It is an interface{}, since our only requirement is that it JSON
	// marshals to a valid list of JSON Patch operations.
	Patch interface{} `json:"patch"`
}
