package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"

	protocol "github.com/sourcegraph/lsif-protocol"
)

// Symbol roughly follows the structure of LSP SymbolInformation
type Symbol struct {
	// TODO(beyang): rename ID
	Identifier string
	Text       string
	Detail     string
	Kind       protocol.SymbolKind

	Tags      []protocol.SymbolTag
	Locations []SymbolLocation
	Monikers  []MonikerData

	Children []*Symbol
	Root     *Symbol
}

type SymbolLocation struct {
	Path  string
	Range Range
}

const symbolsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:{Symbols}
select dump_id, data FROM lsif_data_symbols WHERE dump_id = %d LIMIT %d
`

func (s *Store) Symbols(ctx context.Context, bundleID int, path string) ([]*Symbol, error) {
	symbols, err := s.scanSymbols(s.Store.Query(ctx, sqlf.Sprintf(symbolsQuery, bundleID, 1000)))
	if err != nil {
		return nil, err
	}
	for _, symbol := range symbols {
		assignRoot(symbol, symbol)
	}
	return symbols, nil
}

func assignRoot(symbol, root *Symbol) {
	symbol.Root = root
	for _, child := range symbol.Children {
		assignRoot(child, root)
	}
}
