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
	// Query represents metadata about the symbol that is being searched for.
	Query SymbolDescriptor `json:"query"`

	// Files is an optional list of files to restrict the search to.
	Files []string `json:"files,omitempty"`
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
// variable, class, interface, etc that has a reference to it. It is up to the
// language server to define the schema of this object.
//
// SymbolDescriptor usually uniquely identifies a symbol, but it is not
// guaranteed to do so.
type SymbolDescriptor map[string]interface{}

// SymbolLocationInformation is the response type for the `textDocument/xdefinition` extension.
type SymbolLocationInformation struct {
	// A concrete location at which the definition is located, if any.
	Location lsp.Location `json:"location,omitempty"`
	// Metadata about the definition.
	Symbol SymbolDescriptor `json:"SymbolDescriptor"`
}

// Contains tells if this SymbolDescriptor fully contains all of the keys and
// values in the other symbol descriptor.
func (s SymbolDescriptor) Contains(other SymbolDescriptor) bool {
	for k, v := range other {
		v2, ok := s[k]
		if !ok || v != v2 {
			return false
		}
	}
	return true
}

// String returns a consistently ordered string representation of the
// SymbolDescriptor. It is useful for testing.
func (s SymbolDescriptor) String() string {
	sm := make(sortedMap, 0, len(s))
	for k, v := range s {
		sm = append(sm, mapValue{key: k, value: v})
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
