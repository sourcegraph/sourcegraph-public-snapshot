package langserver

import (
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
)

// This file contains lspext but redefined to suit go-langserver
// needs. Everything in here should be wire compatible.

// symbolDescriptor is the exact fields go-langserver uses for
// lspext.SymbolDescriptor. It should have the same JSON wire format. We make
// it a struct for both type safety as well as better memory efficiency.
type symbolDescriptor struct {
	Package     string `json:"package"`
	PackageName string `json:"packageName"`
	Recv        string `json:"recv"`
	Name        string `json:"name"`
	ID          string `json:"id"`
	Vendor      bool   `json:"vendor"`
}

// Contains ensures that b is a subset of our symbolDescriptor
func (a *symbolDescriptor) Contains(b lspext.SymbolDescriptor) bool {
	for k, v := range b {
		switch k {
		case "package":
			if s, ok := v.(string); !ok || s != a.Package {
				return false
			}
		case "packageName":
			if s, ok := v.(string); !ok || s != a.PackageName {
				return false
			}
		case "recv":
			if s, ok := v.(string); !ok || s != a.Recv {
				return false
			}
		case "name":
			if s, ok := v.(string); !ok || s != a.Name {
				return false
			}
		case "id":
			if s, ok := v.(string); !ok || s != a.ID {
				return false
			}
		case "vendor":
			if s, ok := v.(bool); !ok || s != a.Vendor {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// symbolLocationInformation is lspext.SymbolLocationInformation, but using
// our custom symbolDescriptor
type symbolLocationInformation struct {
	// A concrete location at which the definition is located, if any.
	Location lsp.Location `json:"location,omitempty"`
	// Metadata about the definition.
	Symbol *symbolDescriptor `json:"symbol"`
}

// referenceInformation is lspext.ReferenceInformation using our custom symbolDescriptor
type referenceInformation struct {
	// Reference is the location in the workspace where the `symbol` has been
	// referenced.
	Reference lsp.Location `json:"reference"`

	// Symbol is metadata information describing the symbol being referenced.
	Symbol *symbolDescriptor `json:"symbol"`
}
