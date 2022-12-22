package lsifstore

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetDefinitionLocations returns the set of locations defining the symbol at the given position.
func (s *store) GetDefinitionLocations(ctx context.Context, bundleID int, path string, line, character, limit, offset int) (_ []shared.Location, _ int, err error) {
	extractor := func(r precise.RangeData) precise.ID { return r.DefinitionResultID }
	return s.getLocations(ctx, extractor, "definition_ranges", extractDefinitionRanges, s.operations.getDefinitions, bundleID, path, line, character, limit, offset)
}

// GetReferenceLocations returns the set of locations referencing the symbol at the given position.
func (s *store) GetReferenceLocations(ctx context.Context, bundleID int, path string, line, character, limit, offset int) (_ []shared.Location, _ int, err error) {
	lsifExtractor := func(r precise.RangeData) precise.ID { return r.ReferenceResultID }
	return s.getLocations(ctx, lsifExtractor, "reference_ranges", extractReferenceRanges, s.operations.getReferences, bundleID, path, line, character, limit, offset)
}

// GetImplementationLocations returns the set of locations implementing the symbol at the given position.
func (s *store) GetImplementationLocations(ctx context.Context, bundleID int, path string, line, character, limit, offset int) (_ []shared.Location, _ int, err error) {
	extractor := func(r precise.RangeData) precise.ID { return r.ImplementationResultID }
	return s.getLocations(ctx, extractor, "implementation_ranges", extractImplementationRanges, s.operations.getImplementations, bundleID, path, line, character, limit, offset)
}

func (s *store) getLocations(
	ctx context.Context,
	lsifExtractor func(precise.RangeData) precise.ID,
	scipFieldName string,
	scipExtractor func(*scip.Document, *scip.Occurrence) []*scip.Range,
	operation *observation.Operation,
	bundleID int,
	path string,
	line, character, limit, offset int,
) (_ []shared.Location, _ int, err error) {
	ctx, trace, endObservation := operation.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
		locationsDocumentQuery,
		bundleID,
		path,
		bundleID,
		path,
	)))
	if err != nil || !exists {
		return nil, 0, err
	}

	if documentData.SCIPData != nil {
		trace.AddEvent("SCIPData", attribute.Int("numOccurrences", len(documentData.SCIPData.Occurrences)))
		occurrences := types.FindOccurrences(documentData.SCIPData.Occurrences, int32(line), int32(character))
		trace.AddEvent("FindOccurences", attribute.Int("numIntersectingOccurrences", len(occurrences)))

		for _, occurrence := range occurrences {
			var locations []shared.Location
			if ranges := scipExtractor(documentData.SCIPData, occurrence); len(ranges) != 0 {
				locations = append(locations, convertSCIPRangesToLocations(ranges, bundleID, path)...)
			}

			if occurrence.Symbol != "" && !scip.IsLocalSymbol(occurrence.Symbol) {
				monikerLocations, err := s.scanQualifiedMonikerLocations(s.db.Query(ctx, sqlf.Sprintf(
					locationsSymbolSearchQuery,
					pq.Array([]string{occurrence.Symbol}),
					pq.Array([]int{bundleID}),
					sqlf.Sprintf(scipFieldName),
					bundleID,
					path,
					sqlf.Sprintf(scipFieldName),
				)))
				if err != nil {
					return nil, 0, err
				}
				for _, monikerLocation := range monikerLocations {
					for _, row := range monikerLocation.Locations {
						locations = append(locations, shared.Location{
							DumpID: monikerLocation.DumpID,
							Path:   row.URI,
							Range:  newRange(row.StartLine, row.StartCharacter, row.EndLine, row.EndCharacter),
						})
					}
				}
			}

			if len(locations) > 0 {
				totalCount := len(locations)

				if offset < len(locations) {
					locations = locations[offset:]
				} else {
					locations = []shared.Location{}
				}

				if len(locations) > limit {
					locations = locations[:limit]
				}

				return locations, totalCount, nil
			}
		}

		return nil, 0, nil
	}

	trace.AddEvent("LSIF Data ranges", attribute.Int("numRanges", len(documentData.LSIFData.Ranges)))
	ranges := precise.FindRanges(documentData.LSIFData.Ranges, line, character)
	trace.AddEvent("FindRanges", attribute.Int("numIntersectingRanges", len(ranges)))

	orderedResultIDs := extractResultIDs(ranges, lsifExtractor)
	locationsMap, totalCount, err := s.locations(ctx, bundleID, orderedResultIDs, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	trace.AddEvent("locations", attribute.Int("totalCount", totalCount))

	locations := make([]shared.Location, 0, limit)
	for _, resultID := range orderedResultIDs {
		locations = append(locations, locationsMap[resultID]...)
	}

	return locations, totalCount, nil
}

const locationsDocumentQuery = `
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
		NULL AS monikers,
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

const locationsSymbolSearchQuery = `
WITH RECURSIVE
` + symbolIDsCTEs + `
SELECT
	ss.upload_id,
	'' AS scheme,
	'' AS identifier,
	NULL AS data,
	ss.%s,
	sid.document_path
FROM codeintel_scip_symbols ss
JOIN codeintel_scip_document_lookup sid ON sid.id = ss.document_lookup_id
JOIN matching_symbol_names msn ON msn.id = ss.symbol_id
WHERE
	ss.upload_id = %s AND
	sid.document_path != %s AND
	ss.%s IS NOT NULL
`

// locations queries the locations associated with the given definition or reference identifiers. This
// method returns a map from result set identifiers to another map from document paths to locations
// within that document, as well as a total count of locations within the map.
func (s *store) locations(ctx context.Context, bundleID int, ids []precise.ID, limit, offset int) (_ map[precise.ID][]shared.Location, _ int, err error) {
	ctx, trace, endObservation := s.operations.locations.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.Int("numIDs", len(ids)),
		log.String("ids", idsToString(ids)),
		log.Int("limit", limit),
		log.Int("offset", offset),
	}})
	defer endObservation(1, observation.Args{})

	if len(ids) == 0 {
		return nil, 0, nil
	}

	// Get the list of indexes we need to read in order to find each result set identifier
	indexes, err := s.translateIDsToResultChunkIndexes(ctx, bundleID, ids)
	if err != nil {
		return nil, 0, err
	}
	trace.AddEvent("indexes",
		attribute.Int("numIndexes", len(indexes)),
		attribute.String("indexes", intsToString(indexes)),
	)

	// Read the result sets and construct the set of documents we need to open to resolve range
	// identifiers into actual offsets in a document.
	rangeIDsByResultID, totalCount, err := s.readLocationsFromResultChunks(ctx, bundleID, ids, indexes, "")
	if err != nil {
		return nil, 0, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("totalCount", totalCount))

	// Filter out all data in rangeIDsByResultID that falls outside of the current page. This
	// also returns the set of paths for documents we will need to fetch to resolve the results
	// of the current page.
	rangeIDsByResultID, paths := limitResultMap(ids, rangeIDsByResultID, limit, offset)
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("numPaths", len(paths)),
		attribute.String("paths", strings.Join(paths, ", ")))

	// Hydrate the locations result set by replacing range ids with their actual data from their
	// containing document. This refines the map constructed in the previous step.
	locationsByResultID, _, err := s.readRangesFromDocuments(ctx, bundleID, ids, paths, rangeIDsByResultID, trace)
	if err != nil {
		return nil, 0, err
	}

	return locationsByResultID, totalCount, nil
}

// ErrNoMetadata occurs if we can't determine the number of result chunks for an index.
var ErrNoMetadata = errors.New("no rows in meta table")

// translateIDsToResultChunkIndexes converts a set of result set identifiers within a given bundle into
// a deduplicated and sorted set of result chunk indexes that compoletely cover those identifiers.
func (s *store) translateIDsToResultChunkIndexes(ctx context.Context, bundleID int, ids []precise.ID) ([]int, error) {
	// Mapping ids to result chunk indexes relies on the number of total result chunks written during
	// processing so that we can hash identifiers to their parent result chunk in the same deterministic
	// way.
	numResultChunks, exists, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(translateIDsToResultChunkIndexesQuery, bundleID)))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNoMetadata
	}

	resultChunkIndexMap := map[int]struct{}{}
	for _, id := range ids {
		resultChunkIndexMap[precise.HashKey(id, numResultChunks)] = struct{}{}
	}

	indexes := make([]int, 0, len(resultChunkIndexMap))
	for index := range resultChunkIndexMap {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)

	return indexes, nil
}

const translateIDsToResultChunkIndexesQuery = `
SELECT num_result_chunks FROM lsif_data_metadata WHERE dump_id = %s
`

// resultChunkBatchSize is the maximum number of result chunks we will query at once to resolve a single
// locations request.
const resultChunkBatchSize = 50

// readLocationsFromResultChunks reads the given result chunk indexes for a given bundle. This method returns
// a map from documents to range identifiers that compose each of the given input result set identifiers. If
// a non-empty target path is supplied, then any range falling outside that document path will be omitted from
// the output.
func (s *store) readLocationsFromResultChunks(ctx context.Context, bundleID int, ids []precise.ID, indexes []int, targetPath string) (map[precise.ID]map[string][]precise.ID, int, error) {
	totalCount := 0
	rangeIDsByResultID := make(map[precise.ID]map[string][]precise.ID, len(ids))

	// In order to limit the number of parameters we send to Postgres in the result chunk
	// fetch query, we process the indexes in chunks of maximum size. This will also ensure
	// that Postgres will not have to load an unbounded number of compressed result chunk
	// payloads into memory in order to handle the query.

	for len(indexes) > 0 {
		var batch []int
		if len(indexes) <= resultChunkBatchSize {
			batch, indexes = indexes, nil
		} else {
			batch, indexes = indexes[:resultChunkBatchSize], indexes[resultChunkBatchSize:]
		}

		indexQueries := make([]*sqlf.Query, 0, len(batch))
		for _, index := range batch {
			indexQueries = append(indexQueries, sqlf.Sprintf("%s", index))
		}
		visitResultChunks := s.makeResultChunkVisitor(s.db.Query(ctx, sqlf.Sprintf(
			readLocationsFromResultChunksQuery,
			bundleID,
			sqlf.Join(indexQueries, ","),
		)))

		if err := visitResultChunks(func(index int, resultChunkData precise.ResultChunkData) {
			for _, id := range ids {
				documentIDRangeIDs, exists := resultChunkData.DocumentIDRangeIDs[id]
				if !exists {
					continue
				}

				rangeIDsByDocument := make(map[string][]precise.ID, len(documentIDRangeIDs))
				for _, documentIDRangeID := range documentIDRangeIDs {
					if path, ok := resultChunkData.DocumentPaths[documentIDRangeID.DocumentID]; ok {
						if targetPath != "" && path != targetPath {
							continue
						}

						totalCount++
						rangeIDsByDocument[path] = append(rangeIDsByDocument[path], documentIDRangeID.RangeID)
					}
				}
				rangeIDsByResultID[id] = rangeIDsByDocument
			}
		}); err != nil {
			return nil, totalCount, err
		}
	}

	return rangeIDsByResultID, totalCount, nil
}

const readLocationsFromResultChunksQuery = `
SELECT idx, data FROM lsif_data_result_chunks WHERE dump_id = %s AND idx IN (%s)
`

// documentBatchSize is the maximum number of documents we will query at once to resolve a single locations request.
const documentBatchSize = 50

// readRangesFromDocuments extracts range data from the documents with the given paths. This method returns a map from
// result set identifiers to the set of locations composing that result set. The output resolves the missing data given
// via the rangeIDsByResultID parameter. This method also returns a total count of ranges in the result set.
func (s *store) readRangesFromDocuments(ctx context.Context, bundleID int, ids []precise.ID, paths []string, rangeIDsByResultID map[precise.ID]map[string][]precise.ID, trace observation.TraceLogger) (map[precise.ID][]shared.Location, int, error) {
	totalCount := 0
	locationsByResultID := make(map[precise.ID][]shared.Location, len(ids))

	// In order to limit the number of parameters we send to Postgres in the document
	// fetch query, we process the paths in chunks of maximum size. This will also ensure
	// that Postgres will not have to load an unbounded number of compressed document data
	// payloads into memory in order to handle the query.

	for len(paths) > 0 {
		var batch []string
		if len(paths) <= documentBatchSize {
			batch, paths = paths, nil
		} else {
			batch, paths = paths[:documentBatchSize], paths[documentBatchSize:]
		}

		visitDocuments := s.makeDocumentVisitor(func(path string, document precise.DocumentData) {
			totalCount += s.readRangesFromDocument(bundleID, rangeIDsByResultID, locationsByResultID, path, document, trace)
		})

		pathQueries := make([]*sqlf.Query, 0, len(batch))
		for _, path := range batch {
			pathQueries = append(pathQueries, sqlf.Sprintf("%s", path))
		}
		if err := visitDocuments(s.db.Query(ctx, sqlf.Sprintf(readRangesFromDocumentsQuery, bundleID, sqlf.Join(pathQueries, ",")))); err != nil {
			return nil, 0, err
		}
	}

	return locationsByResultID, totalCount, nil
}

const readRangesFromDocumentsQuery = `
SELECT
	dump_id,
	path,
	data,
	ranges,
	NULL AS hovers,
	NULL AS monikers,
	NULL AS packages,
	NULL AS diagnostics,
	NULL AS scip_document
FROM
	lsif_data_documents
WHERE
	dump_id = %s AND
	path IN (%s)
`

// readRangesFromDocument extracts range data from the given document. This method populates the given locationsByResultId
// map, which resolves the missing data given via the rangeIDsByResultID parameter. This method returns a total count of
// ranges in the result set.
func (s *store) readRangesFromDocument(bundleID int, rangeIDsByResultID map[precise.ID]map[string][]precise.ID, locationsByResultID map[precise.ID][]shared.Location, path string, document precise.DocumentData, trace observation.TraceLogger) int {
	totalCount := 0
	for id, rangeIDsByPath := range rangeIDsByResultID {
		rangeIDs := rangeIDsByPath[path]
		if len(rangeIDs) == 0 {
			continue
		}

		locations := make([]shared.Location, 0, len(rangeIDs))
		for _, rangeID := range rangeIDs {
			if r, exists := document.Ranges[rangeID]; exists {
				locations = append(locations, shared.Location{
					DumpID: bundleID,
					Path:   path,
					Range:  newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter),
				})
			}
		}
		trace.AddEvent("TODO Domain Owner",
			attribute.String("id", string(id)),
			attribute.String("path", path),
			attribute.Int("numLocationsForIDInPath", len(locations)))

		totalCount += len(locations)
		locationsByResultID[id] = append(locationsByResultID[id], locations...)
		sortLocations(locationsByResultID[id])
	}

	return totalCount
}

func newRange(startLine, startCharacter, endLine, endCharacter int) types.Range {
	return types.Range{
		Start: types.Position{
			Line:      startLine,
			Character: startCharacter,
		},
		End: types.Position{
			Line:      endLine,
			Character: endCharacter,
		},
	}
}

// sortLocations sorts locations by document, then by offset within a document.
func sortLocations(locations []shared.Location) {
	sort.Slice(locations, func(i, j int) bool {
		if locations[i].Path == locations[j].Path {
			return compareBundleRanges(locations[i].Range, locations[j].Range)
		}

		return strings.Compare(locations[i].Path, locations[j].Path) < 0
	})
}

// compareBundleRanges returns true if r1's start position occurs before r2's start position.
func compareBundleRanges(r1, r2 types.Range) bool {
	cmp := r1.Start.Line - r2.Start.Line
	if cmp == 0 {
		cmp = r1.Start.Character - r2.Start.Character
	}

	return cmp < 0
}

// extractResultIDs extracts result identifiers from each range in the given list.
// The output list is relative to the input range list, but with duplicates removed.
func extractResultIDs(ranges []precise.RangeData, fn func(r precise.RangeData) precise.ID) []precise.ID {
	resultIDs := make([]precise.ID, 0, len(ranges))
	resultIDMap := make(map[precise.ID]struct{}, len(ranges))

	for _, r := range ranges {
		resultID := fn(r)

		if _, ok := resultIDMap[resultID]; !ok && resultID != "" {
			resultIDs = append(resultIDs, resultID)
			resultIDMap[resultID] = struct{}{}
		}
	}

	return resultIDs
}

// limitResultMap returns a map symmetric to the given rangeIDsByResultID that includes only the
// location results on the current page specified by limit and offset, as well as a deduplicated
// and sorted list of paths that exist in the second-level of the returned map.
func limitResultMap(ids []precise.ID, rangeIDsByResultID map[precise.ID]map[string][]precise.ID, limit, offset int) (limited map[precise.ID]map[string][]precise.ID, referencedPaths []string) {
	limitedRangeIDsByResultID := make(map[precise.ID]map[string][]precise.ID, len(rangeIDsByResultID))

	// Get a deduplicated and ordered set of paths that exist in the second-level of the given
	// map. Iterating by sorted path names here tends to require fewer documents being opened
	// per page. Alternatively, iterating by result identifier (which we had done previously)
	// can make us open the same document on multiple disjoint pages in the result set.
	paths := pathsFromResultMap(rangeIDsByResultID)

	// We append paths to the following (re-used) slice whenever we add a previously unseen
	// path to the second-level of the returned map.
	filteredPaths := paths[:0]

outer:
	for _, path := range paths {
		for _, id := range ids {
			rangeIDsByDocument, ok := limitedRangeIDsByResultID[id]
			if !ok {
				rangeIDsByDocument = map[string][]precise.ID{}
				limitedRangeIDsByResultID[id] = rangeIDsByDocument
			}

			rangeIDs := rangeIDsByResultID[id][path]

			if offset < len(rangeIDs) {
				// Skip leading portion of document
				rangeIDs = rangeIDs[offset:]
				offset = 0
			} else {
				// Skip entire document
				offset -= len(rangeIDs)
				continue
			}

			if limit < len(rangeIDs) {
				// Consume leading portion of document
				rangeIDs = rangeIDs[:limit]
				limit = 0
			} else {
				// Consume entire document
				limit -= len(rangeIDs)
			}

			// Assign adjusted slice of ranges into map
			rangeIDsByDocument[path] = rangeIDs

			// If we haven't added this path added it to the filtered path set. Since
			// our _outer_ iteration is paths, if it exists in the set it will be the
			// most recent element (inserted when processing same path, previous id).
			if len(filteredPaths) == 0 || filteredPaths[len(filteredPaths)-1] != path {
				filteredPaths = append(filteredPaths, path)
			}

			if limit == 0 {
				// Page cannot fit any more results
				break outer
			}
		}
	}

	return limitedRangeIDsByResultID, filteredPaths
}

// pathsFromResultMap returns a deduplicated and sorted set of document paths present in the given map.
func pathsFromResultMap(rangeIDsByResultID map[precise.ID]map[string][]precise.ID) []string {
	pathMap := map[string]struct{}{}
	for _, rangeIDsByPath := range rangeIDsByResultID {
		for path := range rangeIDsByPath {
			pathMap[path] = struct{}{}
		}
	}

	paths := make([]string, 0, len(pathMap))
	for path := range pathMap {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	return paths
}

func idsToString(vs []precise.ID) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, string(v))
	}

	return strings.Join(strs, ", ")
}

func intsToString(vs []int) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.Itoa(v))
	}

	return strings.Join(strs, ", ")
}

type extractedOccurrenceData struct {
	definitions     []*scip.Range
	references      []*scip.Range
	implementations []*scip.Range
	hoverText       []string
}

func extractDefinitionRanges(document *scip.Document, occurrence *scip.Occurrence) []*scip.Range {
	return extractOccurrenceData(document, occurrence).definitions
}

func extractReferenceRanges(document *scip.Document, occurrence *scip.Occurrence) []*scip.Range {
	return append(extractOccurrenceData(document, occurrence).definitions, extractOccurrenceData(document, occurrence).references...)
}

func extractImplementationRanges(document *scip.Document, occurrence *scip.Occurrence) []*scip.Range {
	return extractOccurrenceData(document, occurrence).implementations
}

func extractHoverData(document *scip.Document, occurrence *scip.Occurrence) []string {
	return extractOccurrenceData(document, occurrence).hoverText
}

func extractOccurrenceData(document *scip.Document, occurrence *scip.Occurrence) extractedOccurrenceData {
	if occurrence.Symbol == "" {
		return extractedOccurrenceData{
			hoverText: occurrence.OverrideDocumentation,
		}
	}

	var (
		hoverText               []string
		definitionSymbol        = occurrence.Symbol
		referencesBySymbol      = map[string]struct{}{}
		implementationsBySymbol = map[string]struct{}{}
	)

	// Extract hover text and relationship data from the symbol information that
	// matches the given occurrence. This will give us additional symbol names that
	// we should include in reference and implementation searches.

	if symbol := types.FindSymbol(document, occurrence.Symbol); symbol != nil {
		hoverText = symbol.Documentation

		for _, rel := range symbol.Relationships {
			if rel.IsDefinition {
				definitionSymbol = rel.Symbol
			}
			if rel.IsReference {
				referencesBySymbol[rel.Symbol] = struct{}{}
			}
			if rel.IsImplementation {
				implementationsBySymbol[rel.Symbol] = struct{}{}
			}
		}
	}

	definitions := []*scip.Range{}
	references := []*scip.Range{}
	implementations := []*scip.Range{}

	// Include original symbol names for reference and implementation search below
	referencesBySymbol[occurrence.Symbol] = struct{}{}
	implementationsBySymbol[occurrence.Symbol] = struct{}{}

	// For each occurrence that references one of the definition, reference, or
	// implementation symbol names, extract and aggregate their source positions.

	for _, occ := range document.Occurrences {
		isDefinition := scip.SymbolRole_Definition.Matches(occ)

		// This occurrence defines this symbol
		if definitionSymbol == occ.Symbol && isDefinition {
			definitions = append(definitions, scip.NewRange(occ.Range))
		}

		// This occurrence references this symbol (or a sibling of it)
		if _, ok := referencesBySymbol[occ.Symbol]; ok && !isDefinition {
			references = append(references, scip.NewRange(occ.Range))
		}

		// Either one of the following are true:
		//
		// (1) The source occurrence is a definition, and this occurrence is
		//     an implementation of the source interface, or
		// (2) The source occurrence is a reference, and this occurrence is a
		//     definition of an conforming interface.
		if _, ok := implementationsBySymbol[occ.Symbol]; ok && (isDefinition || scip.SymbolRole_Definition.Matches(occurrence)) {
			implementations = append(implementations, scip.NewRange(occ.Range))
		}
	}

	// Override symbol documentation with occurrence documentation, if it exists
	if len(occurrence.OverrideDocumentation) != 0 {
		hoverText = occurrence.OverrideDocumentation
	}

	return extractedOccurrenceData{
		definitions:     definitions,
		references:      references,
		implementations: implementations,
		hoverText:       hoverText,
	}
}
