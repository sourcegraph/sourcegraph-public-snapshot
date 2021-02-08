package lsifstore

import (
	"context"
	"fmt"
	"strings"

	"github.com/keegancsmith/sqlf"

	protocol "github.com/sourcegraph/lsif-protocol"
)

// Symbol roughly follows the structure of LSP SymbolInformation
type Symbol struct {
	Identifier string
	Text       string
	Detail     string
	Kind       protocol.SymbolKind

	Tags      []protocol.SymbolTag
	Locations []SymbolLocation
	Monikers  []MonikerData

	Children []*Symbol
}

type SymbolLocation struct {
	Path  string
	Range Range
}

// Symbols returns all LSIF document symbols in documents prefixed by path.
func (s *Store) Symbols(ctx context.Context, bundleID int, path string) ([]*Symbol, error) {
	const limit = 10000
	scannedDocumentData, err := s.scanDocumentData(s.Store.Query(ctx, sqlf.Sprintf(documentsQuery, bundleID, path, limit)))
	if err != nil || len(scannedDocumentData) == 0 {
		return nil, err
	}

	numSymbols := 0
	for _, docData := range scannedDocumentData {
		numSymbols += len(docData.Document.Symbols)
	}

	// Convert to symbols
	symbols := make([]*Symbol, 0, numSymbols)
	for _, docData := range scannedDocumentData {
		for _, symData := range docData.Document.Symbols {
			symbol, err := symbolDataToSymbol(docData.Path, docData.Document, symData)
			if err != nil {
				return nil, err
			}
			symbols = append(symbols, symbol)
		}
	}

	monikers, err := s.readMonikerLocations(ctx, bundleID)
	if err != nil {
		return nil, err
	}
	if err := associateMonikers(symbols, monikers); err != nil {
		return nil, err
	}
	// for _, symbol := range symbols {
	// 	log.Printf("# symbol monikers: %v", symbol.Monikers)
	// }

	// TODO(beyang): coalesce using monikers, rather than symbol text
	type symbolKey struct {
		id   string
		kind protocol.SymbolKind
	}
	coalescedSymbols := make(map[symbolKey]*Symbol)
	for _, symbol := range symbols {
		existing, ok := coalescedSymbols[symbolKey{symbol.Text, symbol.Kind}]
		if ok {
			existing.Locations = append(existing.Locations, symbol.Locations...)
			existing.Children = append(existing.Children, symbol.Children...)
			// TODO(beyang): union tags
			details := make([]string, 0, 2)
			if existing.Detail != "" {
				details = append(details, existing.Detail)
			}
			if symbol.Detail != "" {
				details = append(details, symbol.Detail)
			}
			existing.Detail = strings.Join(details, "\n\n")
		} else {
			coalescedSymbols[symbolKey{symbol.Text, symbol.Kind}] = symbol
		}
	}

	coalescedSymbolSlice := make([]*Symbol, 0, len(coalescedSymbols))
	for _, symbol := range coalescedSymbols {
		coalescedSymbolSlice = append(coalescedSymbolSlice, symbol)
	}

	return coalescedSymbolSlice, nil
}

// NEXT
// - See if ranges in moniker data match up with ranges in symbol data (examine actual data)
// - Factor this out into a separate type
// - Determine how to handle if a moniker is not found for a symbol. Synthetic moniker?
//   -> just need to derive a stable identifier for each symbol (distinct but possibly derived from moniker)

var getPosKey = func(startLine, startCharacter, endLine, endCharacter int) string {
	// return fmt.Sprintf("%d:%d:%d:%d", startLine, startCharacter, endLine, endCharacter)
	return fmt.Sprintf("%d:%d", startLine, startCharacter)
}

func associateMonikers(symbols []*Symbol, monikers []MonikerLocations) error {
	monikersByLoc := make(map[string]map[string][]MonikerData)
	addLoc := func(path string, startLine, startCharacter, endLine, endCharacter int, moniker MonikerData) {
		q, ok := monikersByLoc[path]
		if !ok {
			q = make(map[string][]MonikerData)
			monikersByLoc[path] = q
		}
		posKey := getPosKey(startLine, startCharacter, endLine, endCharacter)
		q[posKey] = append(q[posKey], moniker)
	}
	for _, moniker := range monikers {
		for _, loc := range moniker.Locations {
			addLoc(loc.URI, loc.StartLine, loc.StartCharacter, loc.EndLine, loc.EndCharacter, MonikerData{
				Kind:       "export", // TODO(beyang): record this in data_write.go
				Scheme:     moniker.Scheme,
				Identifier: moniker.Identifier,
			})
		}
	}

	for _, symbol := range symbols {
		associateMonikersForSymbol(symbol, monikersByLoc)
	}
	return nil
}

func associateMonikersForSymbol(symbol *Symbol, monikersByLoc map[string]map[string][]MonikerData) {
	// TODO(beyang): factor out monikersByLoc into separate type
	getLocs := func(path string, rng Range) []MonikerData {
		q, ok := monikersByLoc[path]
		if !ok {
			return nil
		}
		posKey := getPosKey(rng.Start.Line, rng.Start.Character, rng.End.Line, rng.End.Character)
		monikers := q[posKey]
		return monikers
	}
	for _, loc := range symbol.Locations {
		locs := getLocs(loc.Path, loc.Range)
		symbol.Monikers = append(symbol.Monikers, locs...)
	}
	for _, child := range symbol.Children {
		associateMonikersForSymbol(child, monikersByLoc)
	}
}

func (s *Store) readMonikerLocations(ctx context.Context, bundleID int) ([]MonikerLocations, error) {
	rows, err := s.Store.Query(ctx, sqlf.Sprintf(`SELECT scheme, identifier, data FROM lsif_data_definitions WHERE dump_id = %s`, bundleID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var values []MonikerLocations
	for rows.Next() {
		var (
			value MonikerLocations
			data  []byte
		)
		if err := rows.Scan(&value.Scheme, &value.Identifier, &data); err != nil {
			return nil, err
		}
		locations, err := s.serializer.UnmarshalLocations([]byte(data))
		if err != nil {
			return nil, err
		}
		value.Locations = locations
		values = append(values, value)
	}
	return values, nil
}
