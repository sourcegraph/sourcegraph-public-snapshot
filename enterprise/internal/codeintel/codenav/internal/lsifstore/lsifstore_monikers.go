package lsifstore

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// GetMonikersByPosition returns all monikers attached ranges containing the given position. If multiple
// ranges contain the position, then this method will return multiple sets of monikers. Each slice
// of monikers are attached to a single range. The order of the output slice is "outside-in", so that
// the range attached to earlier monikers enclose the range attached to later monikers.
func (s *store) GetMonikersByPosition(ctx context.Context, uploadID int, path string, line, character int) (_ [][]precise.MonikerData, err error) {
	ctx, trace, endObservation := s.operations.getMonikersByPosition.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("uploadID", uploadID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
		monikersDocumentQuery,
		uploadID,
		path,
		uploadID,
		path,
	)))
	if err != nil || !exists {
		return nil, err
	}

	if documentData.SCIPData != nil {
		trace.Log(log.Int("numOccurrences", len(documentData.SCIPData.Occurrences)))
		occurrences := types.FindOccurrences(documentData.SCIPData.Occurrences, int32(line), int32(character))
		trace.Log(log.Int("numIntersectingOccurrences", len(occurrences)))

		// Make lookup map of symbol information by name
		symbolMap := map[string]*scip.SymbolInformation{}
		for _, symbol := range documentData.SCIPData.Symbols {
			symbolMap[symbol.Symbol] = symbol
		}

		monikerData := make([][]precise.MonikerData, 0, len(occurrences))
		for _, occurrence := range occurrences {
			if occurrence.Symbol == "" || scip.IsLocalSymbol(occurrence.Symbol) {
				continue
			}
			symbol, hasSymbol := symbolMap[occurrence.Symbol]

			kind := precise.Import
			if hasSymbol {
				for _, o := range documentData.SCIPData.Occurrences {
					if o.Symbol == occurrence.Symbol {
						// TODO - do we need to check additional documents here?
						if isDefinition := scip.SymbolRole_Definition.Matches(o); isDefinition {
							kind = precise.Export
						}

						break
					}
				}
			}

			moniker, err := symbolNameToQualifiedMoniker(occurrence.Symbol, kind)
			if err != nil {
				return nil, err
			}
			occurrenceMonikers := []precise.MonikerData{moniker}

			if hasSymbol {
				for _, rel := range symbol.Relationships {
					if rel.IsImplementation {
						relatedMoniker, err := symbolNameToQualifiedMoniker(rel.Symbol, precise.Implementation)
						if err != nil {
							return nil, err
						}

						occurrenceMonikers = append(occurrenceMonikers, relatedMoniker)
					}
				}
			}

			monikerData = append(monikerData, occurrenceMonikers)
		}
		trace.Log(log.Int("numMonikers", len(monikerData)))

		return monikerData, nil
	}

	trace.Log(log.Int("numRanges", len(documentData.LSIFData.Ranges)))
	ranges := precise.FindRanges(documentData.LSIFData.Ranges, line, character)
	trace.Log(log.Int("numIntersectingRanges", len(ranges)))

	monikerData := make([][]precise.MonikerData, 0, len(ranges))
	for _, r := range ranges {
		batch := make([]precise.MonikerData, 0, len(r.MonikerIDs))
		for _, monikerID := range r.MonikerIDs {
			if moniker, exists := documentData.LSIFData.Monikers[monikerID]; exists {
				batch = append(batch, moniker)
			}
		}
		trace.Log(log.Int("numMonikersForRange", len(batch)))

		monikerData = append(monikerData, batch)
	}
	trace.Log(log.Int("numMonikers", len(monikerData)))

	return monikerData, nil
}

const monikersDocumentQuery = `
(
	SELECT
		sd.id,
		sid.document_path,
		NULL AS data,
		NULL AS ranges,
		NULL AS hovers,
		NULL AS monikers,
		NULL AS packages,
		NULL AS diagnostics,
		sd.raw_scip_payload AS scip_document
	FROM codeintel_scip_document_lookup sid
	JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
	WHERE
		sid.upload_id = %s AND
		sid.document_path = %s
	LIMIT 1
) UNION (
	SELECT
		dump_id,
		path,
		data,
		ranges,
		NULL AS hovers,
		monikers,
		NULL AS packages,
		NULL AS diagnostics,
		NULL AS scip_document
	FROM
		lsif_data_documents
	WHERE
		dump_id = %s AND
		path = %s
	LIMIT 1
)
`

func symbolNameToQualifiedMoniker(symbolName, kind string) (precise.MonikerData, error) {
	parsedSymbol, err := scip.ParseSymbol(symbolName)
	if err != nil {
		return precise.MonikerData{}, err
	}

	return precise.MonikerData{
		Scheme:     parsedSymbol.Scheme,
		Kind:       kind,
		Identifier: symbolName,
		PackageInformationID: precise.ID(strings.Join([]string{
			"scip",
			// Base64 encoding these components as names converted from LSIF contain colons as part of the
			// general moniker scheme. It's reasonable for manager and names in SCIP-land to also have colons,
			// so we'll just remove the ambiguity from the generated string here.
			base64.RawStdEncoding.EncodeToString([]byte(parsedSymbol.Package.Manager)),
			base64.RawStdEncoding.EncodeToString([]byte(parsedSymbol.Package.Name)),
			base64.RawStdEncoding.EncodeToString([]byte(parsedSymbol.Package.Version)),
		}, ":")),
	}, nil
}

// GetBulkMonikerLocations returns the locations (within one of the given uploads) with an attached moniker
// whose scheme+identifier matches one of the given monikers. This method also returns the size of the
// complete result set to aid in pagination.
func (s *store) GetBulkMonikerLocations(ctx context.Context, tableName string, uploadIDs []int, monikers []precise.MonikerData, limit, offset int) (_ []shared.Location, totalCount int, err error) {
	ctx, trace, endObservation := s.operations.getBulkMonikerResults.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("tableName", tableName),
		log.Int("numUploadIDs", len(uploadIDs)),
		log.String("uploadIDs", intsToString(uploadIDs)),
		log.Int("numMonikers", len(monikers)),
		log.String("monikers", monikersToString(monikers)),
		log.Int("limit", limit),
		log.Int("offset", offset),
	}})
	defer endObservation(1, observation.Args{})

	if len(uploadIDs) == 0 || len(monikers) == 0 {
		return nil, 0, nil
	}

	idQueries := make([]*sqlf.Query, 0, len(uploadIDs))
	for _, id := range uploadIDs {
		idQueries = append(idQueries, sqlf.Sprintf("%s", id))
	}

	monikerQueries := make([]*sqlf.Query, 0, len(monikers))
	for _, arg := range monikers {
		monikerQueries = append(monikerQueries, sqlf.Sprintf("(%s, %s)", arg.Scheme, arg.Identifier))
	}

	symbolNames := make([]string, 0, len(monikers))
	for _, arg := range monikers {
		symbolNames = append(symbolNames, arg.Identifier)
	}

	query := sqlf.Sprintf(
		bulkMonikerResultsQuery,
		pq.Array(symbolNames),
		pq.Array(uploadIDs),
		sqlf.Sprintf(fmt.Sprintf("lsif_data_%s", tableName)),
		sqlf.Join(idQueries, ", "),
		sqlf.Join(monikerQueries, ", "),
		sqlf.Sprintf(fmt.Sprintf("%s_ranges", strings.TrimSuffix(tableName, "s"))),
	)

	locationData, err := s.scanQualifiedMonikerLocations(s.db.Query(ctx, query))
	if err != nil {
		return nil, 0, err
	}

	totalCount = 0
	for _, monikerLocations := range locationData {
		totalCount += len(monikerLocations.Locations)
	}
	trace.Log(
		log.Int("numDumps", len(locationData)),
		log.Int("totalCount", totalCount),
	)

	max := totalCount
	if totalCount > limit {
		max = limit
	}

	locations := make([]shared.Location, 0, max)
outer:
	for _, monikerLocations := range locationData {
		for _, row := range monikerLocations.Locations {
			offset--
			if offset >= 0 {
				continue
			}

			locations = append(locations, shared.Location{
				DumpID: monikerLocations.DumpID,
				Path:   row.URI,
				Range:  newRange(row.StartLine, row.StartCharacter, row.EndLine, row.EndCharacter),
			})

			if len(locations) >= limit {
				break outer
			}
		}
	}
	trace.Log(log.Int("numLocations", len(locations)))

	return locations, totalCount, nil
}

const bulkMonikerResultsQuery = `
WITH RECURSIVE
` + symbolIDsCTEs + `
(
	SELECT
		dump_id,
		scheme,
		identifier,
		data,
		NULL AS scip_payload,
		'' AS scip_uri
	FROM %s
	WHERE
		dump_id IN (%s) AND
		(scheme, identifier) IN (%s)
	ORDER BY
		(dump_id, scheme, identifier)
) UNION ALL (
	SELECT
		ss.upload_id,
		'scip',
		msn.symbol_name,
		NULL,
		%s,
		document_path
	FROM matching_symbol_names msn
	JOIN codeintel_scip_symbols ss ON ss.upload_id = msn.upload_id AND ss.symbol_id = msn.id
	JOIN codeintel_scip_document_lookup dl ON dl.id = ss.document_lookup_id
	ORDER BY
		(ss.upload_id, msn.symbol_name)
)
`

func monikersToString(vs []precise.MonikerData) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, fmt.Sprintf("%s:%s:%s", v.Kind, v.Scheme, v.Identifier))
	}

	return strings.Join(strs, ", ")
}
