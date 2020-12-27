package api

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type ResolvedSymbol struct {
	lsifstore.Symbol
	Children []ResolvedSymbol
	Dump     store.Dump
}

// Symbols returns the symbols defined in the given path prefix.
func (api *CodeIntelAPI) Symbols(ctx context.Context, filters *gql.SymbolFilters, uploadID, limit, offset int) (_ []ResolvedSymbol, _ int, err error) {
	ctx, endObservation := api.operations.symbols.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("uploadID", uploadID),
		log.Int("limit", limit),
		log.Int("offset", offset),
	}})
	defer endObservation(1, observation.Args{})

	dump, exists, err := api.dbStore.GetDumpByID(ctx, uploadID)
	if err != nil {
		return nil, 0, errors.Wrap(err, "store.GetDumpByID")
	}
	if !exists {
		return nil, 0, ErrMissingDump
	}

	symbols, totalCount, err := api.lsifStore.Symbols(ctx, dump.ID, filters, offset, limit)
	if err != nil {
		if err == lsifstore.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, 0, nil
		}
		return nil, 0, errors.Wrap(err, "bundleClient.Symbols")
	}

	return resolveSymbolsWithDump(dump, symbols), totalCount, nil
}

// Symbol returns the
func (api *CodeIntelAPI) Symbol(ctx context.Context, uploadID int, scheme, identifier string) (_ *ResolvedSymbol, _ []int, err error) {
	ctx, endObservation := api.operations.symbols.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("uploadID", uploadID),
		log.String("scheme", scheme),
		log.String("identifier", identifier),
	}})
	defer endObservation(1, observation.Args{})

	dump, exists, err := api.dbStore.GetDumpByID(ctx, uploadID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "store.GetDumpByID")
	}
	if !exists {
		return nil, nil, ErrMissingDump
	}

	root, descentPath, err := api.lsifStore.Symbol(ctx, dump.ID, scheme, identifier)
	if root == nil || err != nil {
		if err == lsifstore.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, nil, nil
		}
		return nil, nil, errors.Wrap(err, "bundleClient.Symbol")
	}

	return &resolveSymbolsWithDump(dump, []lsifstore.Symbol{*root})[0], descentPath, nil
}

func resolveSymbolsWithDump(dump store.Dump, roots []lsifstore.Symbol) []ResolvedSymbol {
	var convertToResolved func(s *lsifstore.Symbol) ResolvedSymbol
	convertToResolved = func(s *lsifstore.Symbol) ResolvedSymbol {
		rs := ResolvedSymbol{
			Dump:   dump,
			Symbol: *s,
		}
		for i := range s.Children {
			rs.Children = append(rs.Children, convertToResolved(&s.Children[i]))
		}
		return rs
	}

	return convertToResolved(&lsifstore.Symbol{Children: roots}).Children
}
