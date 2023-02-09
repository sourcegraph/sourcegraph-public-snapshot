package lsifstore

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type DependencyDescription struct {
	Manager string
	Name    string
	Version string
}

// GetDependencies returns a list of dependencies for the given index.
func (s *store) GetDependencies(ctx context.Context, bundleIDs []int) (_ []DependencyDescription, err error) {
	// ctx, _, endObservation := s.operations.getExists.With(ctx, &err, observation.Args{LogFields: []log.Field{
	// 	log.String("bundleIDs", intsToString(bundleIDs)),
	// }})
	// defer endObservation(1, observation.Args{})

	symbolNames, err := basestore.ScanStrings(s.db.Query(ctx, sqlf.Sprintf(
		dependenciesQuery,
		pq.Array(bundleIDs),
	)))
	if err != nil {
		return nil, err
	}

	symbolsByKey := map[string]*scip.Symbol{}
	for _, symbolName := range symbolNames {
		symbol, err := scip.ParseSymbol(symbolName)
		if err != nil {
			return nil, err
		}

		symbolsByKey[fmt.Sprintf("%s:%s:%s", symbol.Package.Manager, symbol.Package.Name, symbol.Package.Version)] = symbol
	}

	var descriptions []DependencyDescription
	for _, symbol := range symbolsByKey {
		descriptions = append(descriptions, DependencyDescription{
			Manager: symbol.Package.Manager,
			Name:    symbol.Package.Name,
			Version: symbol.Package.Version,
		})
	}
	sort.Slice(descriptions, func(i, j int) bool {
		di := descriptions[i]
		dj := descriptions[j]

		if di.Manager == dj.Manager {
			if di.Name == dj.Name {
				return di.Version < dj.Version
			}

			return di.Name < dj.Name
		}

		return di.Manager < dj.Manager
	})

	return descriptions, nil
}

//
// TODO - build top down instead?

const dependenciesQuery = `
WITH RECURSIVE
all_prefixes(upload_id, id, prefix) AS (
	(
		SELECT
			ssn.upload_id,
			ssn.id,
			ssn.name_segment
		FROM codeintel_scip_symbol_names ssn
		WHERE
			ssn.upload_id = ANY(%s) AND
			ssn.prefix_id IS NULL
	) UNION (
		SELECT
			ssn.upload_id,
			ssn.id,
			mp.prefix || ssn.name_segment
		FROM all_prefixes mp
		JOIN codeintel_scip_symbol_names ssn ON
			ssn.upload_id = mp.upload_id AND
			ssn.prefix_id = mp.id
	)
)

SELECT mp.prefix FROM all_prefixes mp
WHERE NOT EXISTS (
	SELECT 1
	FROM codeintel_scip_symbol_names ssn
	WHERE ssn.upload_id = mp.upload_id AND ssn.prefix_id = mp.id
)
`

type Location struct {
	UploadID int
	Path     string
	Range    types.Range
}

// GetDependencies returns a list of dependencies for the given index.
func (s *store) GetDependencyOccurrences(ctx context.Context, bundleIDs []int, manager, name, version string) (_ []Location, err error) {
	// ctx, _, endObservation := s.operations.getExists.With(ctx, &err, observation.Args{LogFields: []log.Field{
	// 	log.String("bundleIDs", intsToString(bundleIDs)),
	// }})
	// defer endObservation(1, observation.Args{})

	result, err := s.scanLocationsFromRows(s.db.Query(ctx, sqlf.Sprintf(
		dependencyOccurrences,
		pq.Array(bundleIDs),
		fmt.Sprintf("%s . %s %s", scipEscape(manager), scipEscape(name), scipEscape(version)),
	)))
	if err != nil {
		return nil, err
	}

	return result, nil
}

//
// TODO - be chill, dude.

const dependencyOccurrences = `
WITH RECURSIVE
all_prefixes(upload_id, id, prefix) AS (
	(
		SELECT
			ssn.upload_id,
			ssn.id,
			ssn.name_segment
		FROM codeintel_scip_symbol_names ssn
		WHERE
			ssn.upload_id = ANY(%s) AND
			ssn.prefix_id IS NULL
	) UNION (
		SELECT
			ssn.upload_id,
			ssn.id,
			mp.prefix || ssn.name_segment
		FROM all_prefixes mp
		JOIN codeintel_scip_symbol_names ssn ON
			ssn.upload_id = mp.upload_id AND
			ssn.prefix_id = mp.id
	)
)

SELECT
	mp.upload_id,
	dl.document_path,
	mp.prefix AS symbol_name,
	ss.definition_ranges,
	ss.reference_ranges
FROM all_prefixes mp
JOIN codeintel_scip_symbols ss ON ss.upload_id = mp.upload_id AND ss.symbol_id = mp.id
JOIN codeintel_scip_document_lookup dl ON ss.upload_id = dl.upload_id AND dl.id = ss.document_lookup_id
WHERE mp.prefix LIKE %s || '%%'
`

func (s *store) scanLocationsFromRows(rows *sql.Rows, queryErr error) (_ []Location, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var values []Location
	for rows.Next() {
		records, err := s.scanLocationsFromRow(rows)
		if err != nil {
			return nil, err
		}

		values = append(values, records...)
	}

	return values, nil
}

func (s *store) scanLocationsFromRow(rows *sql.Rows) ([]Location, error) {
	var (
		uploadID                                int
		documentPath, symbolName                string
		rawDefinitionRanges, rawReferenceRanges []byte
	)

	if err := rows.Scan(&uploadID, &documentPath, &symbolName, &rawDefinitionRanges, &rawReferenceRanges); err != nil {
		return nil, err
	}

	definitionRanges, err := types.DecodeRanges(rawDefinitionRanges)
	if err != nil {
		return nil, err
	}
	referenceRanges, err := types.DecodeRanges(rawReferenceRanges)
	if err != nil {
		return nil, err
	}

	var locations []Location
	for _, r := range definitionRanges {
		locations = append(locations, Location{
			UploadID: uploadID,
			Path:     documentPath,
			Range:    translateRange(r),
		})
	}
	for _, r := range referenceRanges {
		locations = append(locations, Location{
			UploadID: uploadID,
			Path:     documentPath,
			Range:    translateRange(r),
		})
	}

	return locations, nil
}

func scipEscape(s string) string {
	if s == "" {
		return "gomod" // WTF? - some data is not matched up correctly
	}

	return strings.ReplaceAll(s, " ", "  ")
}
