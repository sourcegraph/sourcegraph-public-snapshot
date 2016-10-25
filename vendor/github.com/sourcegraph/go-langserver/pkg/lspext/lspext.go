package lspext

import "github.com/sourcegraph/go-langserver/pkg/lsp"

// WorkspaceReferenceParams is a parameter literal used in `workspace/reference`
// requests. This is a Sourcegraph extension method to LSP. It is sent from the
// client to the server, and the response type is `[]ReferenceInformation`.
//
// It strictly returns the locations in the workspace at which symbols outside
// of the workspace are referenced. That is:
//
// 	- Excluding any `URI` which is located within the workspace.
// 	- Excluding any `Name` which is unexported or private (i.e. according to
// 	  language semantics).
// 	- Excluding any `Location` which is located in vendored code (e.g.
// 	  `vendor/...` for Go, `node_modules/...` for JS, .tgz NPM packages, or
// 	  .jar files for Java).
//
type WorkspaceReferenceParams struct {
	Limit int `json:"limit"`
}

type ReferenceInformation struct {
	// Location is the location at which Symbol has been referenced.
	Location lsp.Location `json:"location"`

	// Name is the name of the symbol that is being referenced. For example:
	//
	// 	(Go) "ServeHTTP"
	// 	(JS) "render"
	//
	Name string `json:"name"`

	// ContainerName is the container name of the symbol that is being
	// referenced. For example:
	//
	// 	(Go) "Router"
	// 	(JS) "ReactMount"
	//
	ContainerName string `json:"containerName,omitempty"`

	// URI is the URI location of the symbol that is being referenced. If both
	// Name and ContainerName are empty strings, it implies that the URI
	// location is referenced but not a particular symbol (e.g. a Go import
	// statement, a JS require() call, etc).
	URI string `json:"uri"`
}
