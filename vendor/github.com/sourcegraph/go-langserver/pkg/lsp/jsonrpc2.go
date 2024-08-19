package lsp

import (
	"github.com/sourcegraph/go-lsp"
)

// ID represents a JSON-RPC 2.0 request ID, which may be either a
// string or number (or null, which is unsupported).
type ID = lsp.ID
