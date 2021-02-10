package correlation

import (
	"errors"
	"fmt"
	"path"

	protocol "github.com/sourcegraph/lsif-protocol"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

func deriveSymbols(state *State) ([]*lsifstore.Symbol, error) {
	for docID, resultID := range state.TextDocumentDocumentSymbol {
		docURI, ok := state.DocumentData[docID]
		if !ok {
			return nil, fmt.Errorf("Document ID %v doesn't exist in state.DocumentData", docID)
		}
		// NEXT
	}

	// Convert symbols
	// Coalesce top-level symbols
	// Associate monikers based on range

	return nil, nil
}

func convertSymbol(state *State, filename string, parentID string, ds *protocol.RangeBasedDocumentSymbol) (*lsifstore.Symbol, error) {
	rng := state.RangeData[ds.ID]
	if rng.Tag == nil {
		return nil, errors.New("RangeBasedDocumentSymbol range has no tag")
	}

	symbolID := path.Join(parentID, rng.Tag.Text)
	var children []*lsifstore.Symbol
	if ds.Children != nil {
		children = make([]*lsifstore.Symbol, len(ds.Children))
		for i, child := range ds.Children {
			symbolChild, err := convertSymbol(state, filename, symbolID, child)
			if err != nil {
				return nil, err
			}
			children[i] = symbolChild
		}
	}
	return &lsifstore.Symbol{
		Identifier: symbolID,
		Text:       rng.Tag.Text,
		Detail:     rng.Tag.Detail,
		Kind:       rng.Tag.Kind,
		Tags:       rng.Tag.Tags,
		Locations: []lsifstore.SymbolLocation{
			Path: filename,
			Range: lsifstore.Range{
				Start: lsifstore.Position{Line: rng.Start.Line, Character: rng.Start.Character},
				End:   lsifstore.Position{Line: rng.End.Line, Character: rng.End.Character},
			},
		},
		Children: children,
	}
}
