// Package lspx contains extensions to the LSP protocol.
//
// An overview of the different protocol variants:
//
// 	// vanilla LSP
// 	github.com/sourcegraph/sourcegraph-go/pkg/lsp
//
// 	// proxy (http gateway) server LSP extensions
// 	sourcegraph.com/sourcegraph/sourcegraph/xlang
//
// 	// (this package) gbuild/lang server LSP extensions
// 	sourcegraph.com/sourcegraph/sourcegraph/xlang/lspx
//
package lspx

import (
	"time"

	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"
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
