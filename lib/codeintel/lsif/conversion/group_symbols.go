package conversion

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion/datastructures"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
)

func gatherSymbols(ctx context.Context, state *State) chan semantic.SymbolData {
	byID := make(map[int]*semantic.SymbolData, len(state.SymbolData))

	// TODO(sqs): speed this operation up by (eg) adding a DocumentID field to RangeData or another
	// map to State that's the reverse of (State).Contains?
	getDocumentContainingRange := func(rangeID int) int {
		var docID int
		state.Contains.Each(func(key int, set *datastructures.IDSet) {
			if set.Contains(rangeID) {
				docID = key
			}
		})
		if docID == 0 {
			panic("docID is 0")
		}
		return docID
	}

	// Gather symbols defined directly on ranges.
	for id, rng := range state.RangeData {
		if rng.Tag != nil && (rng.Tag.Type == "definition" || rng.Tag.Type == "declaration") {
			docID := getDocumentContainingRange(id)
			uri := state.DocumentData[docID]

			byID[id] = &semantic.SymbolData{
				ID:       uint64(id),
				RangeTag: *rng.Tag,
				Location: semantic.LocationData{
					URI:            uri,
					StartLine:      rng.Start.Line,
					EndLine:        rng.End.Line,
					StartCharacter: rng.Start.Character,
					EndCharacter:   rng.End.Character,
				},
			}
		}
	}

	// // Gather symbols defined by symbol vertices.
	// for id, symbol := range state.SymbolData {
	// 	byID[id] = &lsifstore.SymbolData{
	// 		ID:         uint64(id),
	// 		SymbolData: symbol.SymbolData,
	// 		Locations:  symbol.Locations,
	// 	}
	// }

	// Set children from documentSymbolResults.
	for _, results := range state.DocumentSymbolResults {
		// TODO(sqs): determine document ID
		for _, result := range results {
			// TODO(sqs): remove this vallidation once we validate these in correlateDocumentSymbolResult
			symbol, ok := byID[int(result.ID)]
			if !ok {
				panic("symbol not found")
			}

			for _, child := range result.Children {
				symbol.Children = append(symbol.Children, child.ID)
			}
		}
	}

	// Attach monikers.
	for id, data := range byID {
		if symbolMonikers := state.Monikers.Get(id); symbolMonikers != nil {
			symbolMonikers.Each(func(monikerID int) {
				moniker := state.MonikerData[monikerID]
				data.Monikers = append(data.Monikers, semantic.MonikerData{
					Kind:       moniker.Kind,
					Scheme:     moniker.Scheme,
					Identifier: moniker.Identifier,
				})
			})
		}
	}

	// TODO(sqs): parallelize the sending to this channel
	ch := make(chan semantic.SymbolData)

	go func() {
		defer close(ch)

		for _, data := range byID {
			select {
			case ch <- *data:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}
