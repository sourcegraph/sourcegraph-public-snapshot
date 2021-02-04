package api

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

type ResolvedSymbol struct {
	Dump store.Dump
	*lsifstore.Symbol
}

func (api *CodeIntelAPI) Symbols(ctx context.Context, path string, uploadID int) ([]ResolvedSymbol, error) {
	dump, exists, err := api.dbStore.GetDumpByID(ctx, uploadID)
	if err != nil {
		return nil, errors.Wrap(err, "store.GetDumpByID")
	}
	if !exists {
		return nil, ErrMissingDump
	}

	pathInBundle := strings.TrimPrefix(path, dump.Root)

	symbols, err := api.lsifStore.Symbols(ctx, uploadID, pathInBundle)
	if err != nil || len(symbols) == 0 {
		return nil, err
	}

	resolvedSymbols := make([]ResolvedSymbol, len(symbols))
	for i, symbol := range symbols {
		resolvedSymbols[i] = ResolvedSymbol{Dump: dump, Symbol: symbol}
	}
	return resolvedSymbols, nil
}
