package lsifstore

import (
	"context"
	"database/sql"

	pkgerrors "github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
)

// Symbols returns all symbols (subject to the filters).
func (s *Store) Symbols(ctx context.Context, bundleID int, skip, take int) (_ []SymbolNode, _ int, err error) {
	ctx, endObservation := s.operations.symbols.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.Int("skip", skip),
		log.Int("take", take),
	}})
	defer endObservation(1, observation.Args{})

	symbolDatas, err := s.ReadSymbols(ctx, bundleID)
	if err != nil {
		return nil, 0, pkgerrors.Wrap(err, "store.ReadSymbols")
	}

	rootSymbols := buildSymbolTree(symbolDatas, bundleID)

	// Try to associate a moniker with each symbol.
	allMonikers, err := s.readMonikerLocations(ctx, bundleID, "definitions", skip, take)
	if err != nil {
		return nil, 0, pkgerrors.Wrap(err, "store.ReadDefinitions")
	}
	for i := range rootSymbols {
		WalkSymbolTree(&rootSymbols[i], func(symbol *SymbolNode) {
			associateMoniker(symbol, allMonikers)
		})
	}

	totalCount := len(rootSymbols) // TODO(sqs): doesnt account for skip/take

	return rootSymbols, totalCount, nil
}

// Symbol looks up a symbol by its moniker.
func (s *Store) Symbol(ctx context.Context, bundleID int, scheme, identifier string) (_ *semantic.SymbolData, _ []int, err error) {
	ctx, endObservation := s.operations.symbols.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("scheme", scheme),
		log.String("identifier", identifier),
	}})
	defer endObservation(1, observation.Args{})

	monikers := []semantic.MonikerData{
		{Scheme: scheme, Identifier: identifier},
	}
	locations, _, err := s.BulkMonikerResults(ctx, "definitions", []int{bundleID}, monikers, 10, 0)
	if err != nil {
		return nil, nil, err
	}
	if len(locations) > 0 {
		l := locations[0]
		return &semantic.SymbolData{
			ID:       123,
			RangeTag: protocol.RangeTag{Kind: 4},
			Location: semantic.LocationData{
				URI:            l.Path,
				StartLine:      l.Range.Start.Line,
				StartCharacter: l.Range.Start.Character,
				EndLine:        l.Range.End.Line,
				EndCharacter:   l.Range.End.Character,
			},
			Monikers: monikers,
		}, nil, nil
	}

	rootSymbols, _, err := s.Symbols(ctx, bundleID, 0, 0)
	if err != nil {
		return nil, nil, pkgerrors.Wrap(err, "store.Symbols")
	}

	for _, root := range rootSymbols {
		var theSym *SymbolNode
		treePath, ok := findPathToSymbolInTree(&root, func(symbol *SymbolNode) bool {
			for _, m := range symbol.Monikers {
				if m.Scheme == scheme && m.Identifier == identifier {
					theSym = symbol
					return true
				}
			}
			return false
		})
		if ok {
			return &theSym.SymbolData, treePath, nil
		}
	}

	return nil, nil, nil
}

// TODO(sqs): rename readDefinitionReferences to readDefinitionReferencesForMoniker
func (s *Store) readMonikerLocations(ctx context.Context, bundleID int, tableName string, skip, take int) (_ []semantic.MonikerLocations, err error) {
	scanMonikerLocations := func(rows *sql.Rows, queryErr error) (_ []semantic.MonikerLocations, err error) {
		if queryErr != nil {
			return nil, queryErr
		}
		defer func() { err = basestore.CloseRows(rows, err) }()

		var values []semantic.MonikerLocations
		for rows.Next() {
			var (
				value semantic.MonikerLocations
				data  []byte // raw locations data
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

	if tableName == "definitions" {
		tableName = "lsif_data_definitions"
	} else {
		panic("TODO(sqs)")
	}

	return scanMonikerLocations(s.Store.Query(
		ctx,
		sqlf.Sprintf(
			`SELECT scheme, identifier, data FROM "`+tableName+`" WHERE dump_id = %s AND (identifier LIKE %s) LIMIT %d OFFSET %d`,
			bundleID,
			"%:%", // TODO(sqs): hack, omit "local" monikers from lsif-node with
			take,
			skip,
		),
	))
}

func (s *Store) ReadSymbols(ctx context.Context, bundleID int) (_ []semantic.SymbolData, err error) {
	ctx, endObservation := s.operations.readSymbols.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
	}})
	defer endObservation(1, observation.Args{})

	datas, err := basestore.ScanStrings(s.Store.Query(
		ctx,
		sqlf.Sprintf(
			`SELECT data FROM lsif_data_symbols WHERE dump_id = %s`,
			bundleID,
		),
	))
	if err != nil {
		return nil, err
	}

	symbols := make([]semantic.SymbolData, len(datas))
	for i, data := range datas {
		symbols[i], err = s.serializer.UnmarshalSymbolData([]byte(data))
		if err != nil {
			return nil, err
		}
	}

	return symbols, nil
}

func buildSymbolTree(datas []semantic.SymbolData, dumpID int) (roots []SymbolNode) {
	byID := map[uint64]semantic.SymbolData{}
	nonRoots := make(map[uint64]struct{}, len(datas)) // length guess (most are non-roots)
	for _, data := range datas {
		byID[data.ID] = data
		for _, child := range data.Children {
			nonRoots[child] = struct{}{}
		}
	}

	var newSymbol func(data semantic.SymbolData) SymbolNode
	newSymbol = func(data semantic.SymbolData) SymbolNode {
		symbol := SymbolNode{
			DumpID:     dumpID,
			SymbolData: data,
		}
		for _, child := range data.Children {
			symbol.Children = append(symbol.Children, newSymbol(byID[child]))
		}
		return symbol
	}

	for _, data := range datas {
		if _, isNonRoot := nonRoots[data.ID]; !isNonRoot {
			roots = append(roots, newSymbol(data))
		}
	}

	return roots
}

func WalkSymbolTree(root *SymbolNode, walkFn func(symbol *SymbolNode)) {
	walkFn(root)

	for i := range root.Children {
		WalkSymbolTree(&root.Children[i], walkFn)
	}
}

func findPathToSymbolInTree(root *SymbolNode, matchFn func(symbol *SymbolNode) bool) ([]int, bool) {
	if matchFn(root) {
		return nil, true
	}

	for i := range root.Children {
		path, ok := findPathToSymbolInTree(&root.Children[i], matchFn)
		if ok {
			return append([]int{i}, path...), true
		}
	}

	return nil, false
}

func associateMoniker(symbol *SymbolNode, allMonikers []semantic.MonikerLocations) {
	for _, loc := range []semantic.LocationData{symbol.Location} {
		for _, moniker := range allMonikers {
			for _, monikerLoc := range moniker.Locations {
				if loc.URI == monikerLoc.URI &&
					loc.StartLine == monikerLoc.StartLine &&
					loc.StartCharacter == monikerLoc.StartCharacter &&
					loc.EndLine == monikerLoc.EndLine &&
					loc.EndCharacter == monikerLoc.EndCharacter {
					symbol.Monikers = append(symbol.Monikers, semantic.MonikerData{
						Kind:       "export",
						Scheme:     moniker.Scheme,
						Identifier: moniker.Identifier,
					})
					break
				}
			}
		}
	}
}

func trimSymbolTree(roots *[]SymbolNode, keepFn func(symbol *SymbolNode) bool) {
	keep := (*roots)[:0]
	for i := range *roots {
		if keepFn(&(*roots)[i]) {
			trimSymbolTree(&(*roots)[i].Children, keepFn)
			keep = append(keep, (*roots)[i])
		}
	}
	*roots = keep
}
