// Package lspx contains extensions to the LSP protocol.
//
// An overview of the different protocol variants:
//
// 	// vanilla LSP
// 	github.com/sourcegraph/go-langserver/pkg/lsp
//
// 	// proxy (http gateway) server LSP extensions
// 	sourcegraph.com/sourcegraph/sourcegraph/xlang
//
// 	// (this package) build/lang server LSP extensions
// 	sourcegraph.com/sourcegraph/sourcegraph/xlang/lspx
//
package lspext

import (
	"reflect"
	"time"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

// DependencyReference represents a reference to a dependency. []DependencyReference
// is the return type for the build server workspace/xdependencies method.
type DependencyReference struct {
	// Attributes describing the dependency that is being referenced. It is up
	// to the language server to define the schema of this object.
	Attributes map[string]interface{} `json:"attributes,omitempty"`

	// Hints is treated as an opaque object and passed directly by Sourcegraph
	// into the language server's workspace/xreferences method to help narrow
	// down the search space for the references to the symbol.
	//
	// If a language server emits no hints, Sourcegraph will pass none as a
	// parameter to workspace/xreferences which means it must search the entire
	// repository (workspace) in order to find references. workspace/xdependencies
	// should emit sufficient hints here for making all workspace/xreference
	// queries complete in a reasonable amount of time (less than a few seconds
	// for very large repositories). For example, one may include the
	// containing "package" or other build-system level "code unit". Emitting
	// the exact file is not recommended in general as that would produce more
	// data for little performance gain in most situations.
	Hints map[string]interface{} `json:"hints,omitempty"`
}

// PackageInformation is the metadata associated with a build-system-
// or package-manager-level package. Sometimes, languages have
// abstractions called "packages" as well, but this refers
// specifically to packages as defined by the build system or package
// manager. E.g., Python pip packages (NOT Python language packages or
// modules), Go packages, Maven packages (NOT Java language packages),
// npm modules (NOT JavaScript language modules).  PackageInformation
// includes both attributes of the package itself and attributes of
// the package's dependencies.
type PackageInformation struct {
	// Package is the set of attributes of the package
	Package map[string]interface{} `json:"package,omitempty"`

	// Dependencies is the list of dependency attributes
	Dependencies []DependencyReference `json:"dependencies,omitempty"`
}

// TelemetryEventParams is a telemetry/event message sent from a
// build/lang server back to the proxy. The information here is
// forwarded to our opentracing system.
type TelemetryEventParams struct {
	Op        string            `json:"op"`             // the operation name
	StartTime time.Time         `json:"startTime"`      // when the operation started
	EndTime   time.Time         `json:"endTime"`        // when the operation ended
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

	// Mode is the name of the language. It is used to determine the correct
	// language server to route a request to, and to inform a language server
	// what languages it should contribute.
	Mode string `json:"mode"`
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
		default: // structs with a "URI" field
			rv := reflect.ValueOf(o)
			if rv.Kind() == reflect.Ptr {
				rv = rv.Elem()
			}
			if rv.Kind() == reflect.Struct {
				if fv := rv.FieldByName("URI"); fv.Kind() == reflect.String {
					if collect != nil {
						collect(fv.String())
					}
					if update != nil {
						fv.SetString(update(fv.String()))
					}
				}
				for i := 0; i < rv.NumField(); i++ {
					fv := rv.Field(i)
					if fv.Kind() == reflect.Ptr || fv.Kind() == reflect.Struct || fv.Kind() == reflect.Array {
						walk(fv.Interface())
					}
				}
			}
		}
	}
	walk(o)
}

// TODO(keegancsmith) work out why we have so many custom initializeparams
// that look similiar.

// ClientProxyInitializeParams are sent by the client to the proxy in
// the "initialize" request. It has a non-standard field "mode", which
// is the name of the language (using vscode terminology); "go" or
// "typescript", for example.
type ClientProxyInitializeParams struct {
	lsp.InitializeParams
	InitializationOptions ClientProxyInitializationOptions `json:"initializationOptions"`

	// Mode is DEPRECATED; it was moved to the subfield
	// initializationOptions.Mode. It is still here for backward
	// compatibility until the xlang service is upgraded.
	Mode string `json:"mode,omitempty"`
}

// ClientProxyInitializationOptions is the "initializationOptions"
// field of the "initialize" request params sent from the client to
// the LSP client proxy.
type ClientProxyInitializationOptions struct {
	Mode string `json:"mode"`

	// Session, if set, causes this session to be isolated from other
	// LSP sessions using the same workspace and mode. See
	// (contextID).session for more information.
	Session string `json:"session,omitempty"`
}
