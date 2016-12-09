package lspext

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

// WorkspaceReferencesParams is parameters for the `workspace/xreferences` extension
//
// See: https://github.com/sourcegraph/language-server-protocol/blob/master/extension-workspace-reference.md
//
type WorkspaceReferencesParams struct {
}

// ReferenceInformation represents information about a reference to programming
// constructs like variables, classes, interfaces etc.
type ReferenceInformation struct {
	// Reference is the location in the workspace where the `symbol` has been
	// referenced.
	Reference lsp.Location `json:"reference"`

	// Symbol is metadata information describing the symbol being referenced.
	Symbol SymbolDescriptor `json:"symbol"`
}

// SymbolDescriptor represents information about a programming construct like a
// variable, class, interface etc that has a reference to it. Effectively, it
// contains data similar to SymbolInformation except all fields are optional.
//
// SymbolDescriptor usually uniquely identifies a symbol, but it is not
// guaranteed to do so.
type SymbolDescriptor struct {
	// Name of this symbol (same as `SymbolInformation.name`).
	Name string `json:"name,omitempty"`

	// Kind of this symbol (same as `SymbolInformation.kind`).
	Kind lsp.SymbolKind `json:"kind,omitempty"`

	// URI of this symbol (same as `SymbolInformation.location.uri`).
	URI string `json:"uri,omitempty"`

	// ContainerName represents the name of the symbol containing this symbol
	// (same as `SymbolInformation.containerName`).
	ContainerName string `json:"containerName,omitempty"`

	// Whether or not the symbol is defined inside of "vendored" code. In Go, for
	// example, this means that an external dependency was copied to a subdirectory
	// named `vendor`. The exact definition of vendor depends on the language,
	// but it is generally understood to mean "code that was copied from its
	// original source and now lives in our project directly".
	Vendor bool `json:"vendor,omitempty"`

	// Attributes describing the symbol that is being referenced. It is up to
	// the language server to define what exact data this object contains.
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// LocationInformation is the response type for the `textDocument/xdefinition` extension.
type LocationInformation struct {
	// A concrete location at which the definition is located, if any.
	Location lsp.Location `json:"location,omitempty"`
	// Metadata about the definition.
	Symbol SymbolDescriptor `json:"SymbolDescriptor"`
}

// String returns a consistently ordered string representation of the
// SymbolDescriptor. It is useful for testing.
func (s SymbolDescriptor) String() string {
	sm := make(sortedMap, 0, len(s.Attributes))
	for k, v := range s.Attributes {
		sm = append(sm, mapValue{key: "attr_" + k, value: v})
	}
	stdfield := func(k, v string) {
		if v != "" {
			sm = append(sm, mapValue{key: k, value: v})
		}
	}
	stdfield("name", s.Name)
	stdfield("kind", s.Kind.String())
	stdfield("uri", s.URI)
	stdfield("containerName", s.ContainerName)
	if s.Vendor {
		stdfield("vendor", "true")
	}
	sort.Sort(sm)
	var str string
	for _, v := range sm {
		str += fmt.Sprintf("%s:%v ", v.key, v.value)
	}
	return strings.TrimSpace(str)
}

type mapValue struct {
	key   string
	value interface{}
}

type sortedMap []mapValue

func (s sortedMap) Len() int           { return len(s) }
func (s sortedMap) Less(i, j int) bool { return s[i].key < s[j].key }
func (s sortedMap) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
