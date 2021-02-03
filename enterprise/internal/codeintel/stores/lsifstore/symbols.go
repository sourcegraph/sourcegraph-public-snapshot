package lsifstore

import protocol "github.com/sourcegraph/lsif-protocol"

// Symbol roughly follows the structure of LSP SymbolInformation
type Symbol struct {
	Text   string
	Detail string
	Kind   protocol.SymbolKind

	Tags      []protocol.SymbolTag
	Locations []SymbolLocation
	// Monikers  []protocol.Moniker

	Children []*Symbol
}

type SymbolLocation struct {
	Path  string
	Range Range
}
