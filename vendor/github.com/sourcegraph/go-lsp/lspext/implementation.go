package lspext

import "github.com/sourcegraph/go-lsp"

// ImplementationLocation is a superset of lsp.Location with additional Go-specific information
// about the implementation.
type ImplementationLocation struct {
	lsp.Location // the location of the implementation

	// Type is the type of implementation relationship described by this location.
	//
	// If a type T was queried, the set of possible values are:
	//
	// - "to": named or ptr-to-named types assignable to interface T
	// - "from": named interfaces assignable from T (or only from *T if Ptr == true)
	//
	// If a method M on type T was queried, the same set of values above is used, except they refer
	// to methods on the described type (not the described type itself).
	//
	// (This type description is taken from golang.org/x/tools/cmd/guru.)
	Type string `json:"type,omitempty"`

	// Ptr is whether this implementation location is only assignable from a pointer *T (where T is
	// the queried type).
	Ptr bool `json:"ptr,omitempty"`

	// Method is whether a method was queried. If so, then the implementation locations refer to the
	// corresponding methods on the types found by the implementation query (not the types
	// themselves).
	Method bool `json:"method"`
}
