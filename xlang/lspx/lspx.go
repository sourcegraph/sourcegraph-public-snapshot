// Package lspx contains extensions to the LSP protocol.
//
// In addition to the types described here, Sourcegraph extends the
// LSP protocol in the following ways.
//
// 1. If the client sends a textDocument/didOpen message as a request
//    (with a nonempty request "id" field), then the server must treat
//    it as a request (even though the LSP spec says this method is a
//    notification). The server must reply with an empty result (if
//    there is no error) or an error (if there is an error).
//
//    This behavior is mandated by the JSON-RPC 2.0 spec; it
//    would be invalid for the server to not reply.
//
//    TODO(sqs): A valid LSP and JSON-RPC 2.0 server could, however,
//    refuse to treat textDocument/didOpen as a request (as opposed to
//    a notification) and return an error. Therefore we should rely on the
//    server's published capabilities to know if the server can reply
//    to textDocument/didOpen requests.
package lspx

import (
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

// TelemetryEventParams is a telemetry/event message sent from a
// build/lang server back to the proxy. The information here is
// forwarded to our opentracing system.
type TelemetryEventParams struct {
	Op        string            `json:"op"`             // the operation name
	StartTime time.Time         `json:"startTime"`      // when the operation started
	EndTime   time.Time         `json:"startTime"`      // when the operation ended
	Tags      map[string]string `json:"tags,omitempty"` // other metadata
}

type InitializeParams struct {
	lsp.InitializeParams

	// OriginalRootPath is the original rootPath for this LSP session,
	// before any path rewriting occurred. It is typically a Git clone
	// URL of the form
	// "git://github.com/facebook/react.git?rev=master#lib".
	//
	// The Go lang/build server uses this to infer the import path
	// root (and directory structure) to use for a workspace.
	OriginalRootPath string `json:"originalRootPath"`
}

// WalkURIFields walks the LSP params/result object for fields
// containing document URIs.
//
// If collect is non-nil, it calls collect(uri) for every URI
// encountered. Callers can use this to collect a list of all document
// URIs referenced in the params/result.
//
// If update is non-nil, it updates all document URIs in an LSP
// params/result with the value of f(existingURI). Callers can use
// this to rewrite paths in the params/result.
//
// TODO(sqs): does not support WorkspaceEdit (with a field whose
// TypeScript type is {[uri: string]: TextEdit[]}.
func WalkURIFields(o interface{}, collect func(string), update func(string) string) {
	var walk func(o interface{})
	walk = func(o interface{}) {
		switch o := o.(type) {
		case map[string]interface{}:
			for k, v := range o { // Location, TextDocumentIdentifier, TextDocumentItem, etc.
				if k == "uri" {
					if s, ok := v.(string); ok {
						if collect != nil {
							collect(s)
						}
						if update != nil {
							o[k] = update(s)
						}
						continue
					}
				}
				walk(v)
			}
		case []interface{}: // Location[]
			for _, v := range o {
				walk(v)
			}
		}
	}
	walk(o)
}
