package api

import (
	"context"
	"log"
	"strings"

	"github.com/pkg/errors"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

type ResolvedSymbol struct {
	Dump store.Dump
}

type SymbolLocation struct {
	Path  string
	Range lsifstore.Range
}

func (api *CodeIntelAPI) Symbols(ctx context.Context, path string, uploadID int) ([]ResolvedSymbol, error) {
	// MARK
	api.dbStore.GetDumpByID(ctx, uploadID)
	// log.Printf("### uploadID: %v", uploadID)
	// debug.PrintStack()

	dump, exists, err := api.dbStore.GetDumpByID(ctx, uploadID)
	if err != nil {
		return nil, errors.Wrap(err, "store.GetDumpByID")
	}
	if !exists {
		return nil, ErrMissingDump
	}

	pathInBundle := strings.TrimPrefix(path, dump.Root)

	rawSymbolData, err := api.lsifStore.Symbols(ctx, uploadID, pathInBundle)
	if err != nil || len(rawSymbolData) == 0 {
		return nil, err
	}
	log.Printf("# rawSymbolData: %v", rawSymbolData)

	// NEXT: fetch monikers

	// Coalesce and combine into symbols
	symbols := make([]ResolvedSymbol, 0, len(rawSymbolData))
	for _, rawSymbol := range rawSymbolData {
		log.Printf("# rawSymbol: %v", rawSymbol)
		// symbols[i] = ResolvedSymbol{
		// }
	}

	log.Printf("# symbols: %+v", symbols)
	return nil, nil
}
