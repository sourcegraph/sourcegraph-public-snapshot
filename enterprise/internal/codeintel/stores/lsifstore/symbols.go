package lsifstore

import (
	protocol "github.com/sourcegraph/lsif-protocol"
)

// Symbol roughly follows the structure of LSP SymbolInformation
type Symbol struct {
	Text   string
	Detail string
	Kind   protocol.SymbolKind

	Tags      []lsp.SymbolTag
	Monikers  []lsp.Moniker
	Locations []SymbolLocation

	SymbolData []DocumentSymbolData
	Children   []*Symbol
}
