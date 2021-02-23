package lsifstore

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// ErrNotFound occurs when data does not exist for a requested bundle.
var ErrNotFound = errors.New("data does not exist")

// Exists determines if the path exists in the database.
func (s *Store) Exists(ctx context.Context, bundleID int, path string) (_ bool, err error) {
	ctx, endObservation := s.operations.exists.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
	}})
	defer endObservation(1, observation.Args{})

	_, exists, err := basestore.ScanFirstString(s.Store.Query(ctx, sqlf.Sprintf(existsQuery, bundleID, path)))
	return exists, err
}

const existsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:Exists
SELECT path FROM lsif_data_documents WHERE dump_id = %s AND path = %s LIMIT 1
`

// Ranges returns definition, reference, and hover data for each range within the given span of lines.
func (s *Store) Ranges(ctx context.Context, bundleID int, path string, startLine, endLine int) (_ []CodeIntelligenceRange, err error) {
	ctx, traceLog, endObservation := s.operations.ranges.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("startLine", startLine),
		log.Int("endLine", endLine),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(documentQuery, bundleID, path)))
	if err != nil || !exists {
		return nil, err
	}
	traceLog(log.Int("numRanges", len(documentData.Document.Ranges)))

	ranges := map[ID]RangeData{}
	for id, r := range documentData.Document.Ranges {
		if RangeIntersectsSpan(r, startLine, endLine) {
			ranges[id] = r
		}
	}
	traceLog(log.Int("numIntersectingRanges", len(ranges)))

	resultIDMap := make(map[ID]struct{}, 2*len(ranges))
	for _, r := range ranges {
		if r.DefinitionResultID != "" {
			resultIDMap[r.DefinitionResultID] = struct{}{}
		}
		if r.ReferenceResultID != "" {
			resultIDMap[r.ReferenceResultID] = struct{}{}
		}
	}

	resultIDs := make([]ID, 0, len(resultIDMap))
	for id := range resultIDMap {
		resultIDs = append(resultIDs, id)
	}

	locations, err := s.locations(ctx, bundleID, resultIDs)
	if err != nil {
		return nil, err
	}

	codeintelRanges := make([]CodeIntelligenceRange, 0, len(ranges))
	for _, r := range ranges {
		var hoverText string
		if r.HoverResultID != "" {
			if text, exists := documentData.Document.HoverResults[r.HoverResultID]; exists {
				hoverText = text
			}
		}

		// Return only references that are in the same file. Otherwise this set
		// gets very big and such results are of limited use to consumers such as
		// the code intel extensions, which only use references for highlighting
		// uses of an identifier within the same file.
		fileLocalReferences := make([]Location, 0, len(locations[r.ReferenceResultID]))
		for _, r := range locations[r.ReferenceResultID] {
			if r.Path == path {
				fileLocalReferences = append(fileLocalReferences, r)
			}
		}

		codeintelRanges = append(codeintelRanges, CodeIntelligenceRange{
			Range:       newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter),
			Definitions: locations[r.DefinitionResultID],
			References:  fileLocalReferences,
			HoverText:   hoverText,
		})
	}

	sort.Slice(codeintelRanges, func(i, j int) bool {
		return compareBundleRanges(codeintelRanges[i].Range, codeintelRanges[j].Range)
	})

	return codeintelRanges, nil
}

const documentQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:{Ranges,Definitions,References,Hover,MonikersByPosition}
SELECT dump_id, path, data FROM lsif_data_documents WHERE dump_id = %s AND path = %s LIMIT 1
`

// Definitions returns the set of locations defining the symbol at the given position.
func (s *Store) Definitions(ctx context.Context, bundleID int, path string, line, character, limit, offset int) (_ []Location, _ int, err error) {
	extractor := func(r RangeData) ID { return r.DefinitionResultID }
	operation := s.operations.definitions
	return s.definitionsReferences(ctx, extractor, operation, bundleID, path, line, character, limit, offset)
}

// References returns the set of locations referencing the symbol at the given position.
func (s *Store) References(ctx context.Context, bundleID int, path string, line, character, limit, offset int) (_ []Location, _ int, err error) {
	extractor := func(r RangeData) ID { return r.ReferenceResultID }
	operation := s.operations.references
	return s.definitionsReferences(ctx, extractor, operation, bundleID, path, line, character, limit, offset)
}

func (s *Store) definitionsReferences(ctx context.Context, extractor func(r RangeData) ID, operation *observation.Operation, bundleID int, path string, line, character, limit, offset int) (_ []Location, _ int, err error) {
	ctx, traceLog, endObservation := operation.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(documentQuery, bundleID, path)))
	if err != nil || !exists {
		return nil, 0, err
	}

	traceLog(log.Int("numRanges", len(documentData.Document.Ranges)))
	ranges := FindRanges(documentData.Document.Ranges, line, character)
	traceLog(log.Int("numIntersectingRanges", len(ranges)))

	orderedResultIDs := extractResultIDs(ranges, extractor)
	locationsMap, err := s.locations(ctx, bundleID, orderedResultIDs)
	if err != nil {
		return nil, 0, err
	}

	totalCount := 0
	for _, locations := range locationsMap {
		totalCount += len(locations)
	}
	traceLog(log.Int("totalCount", totalCount))

	max := totalCount
	if totalCount > limit {
		max = limit
	}

	locations := make([]Location, 0, max)
outer:
	for _, resultID := range orderedResultIDs {
		for _, location := range locationsMap[resultID] {
			offset--
			if offset >= 0 {
				continue
			}

			locations = append(locations, location)
			if len(locations) >= limit {
				break outer
			}
		}
	}

	return locations, totalCount, nil
}

// Hover returns the hover text of the symbol at the given position.
func (s *Store) Hover(ctx context.Context, bundleID int, path string, line, character int) (_ string, _ Range, _ bool, err error) {
	ctx, traceLog, endObservation := s.operations.hover.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(documentQuery, bundleID, path)))
	if err != nil || !exists {
		return "", Range{}, false, err
	}

	traceLog(log.Int("numRanges", len(documentData.Document.Ranges)))
	ranges := FindRanges(documentData.Document.Ranges, line, character)
	traceLog(log.Int("numIntersectingRanges", len(ranges)))

	for _, r := range ranges {
		if text, ok := documentData.Document.HoverResults[r.HoverResultID]; ok {
			return text, newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter), true, nil
		}
	}

	return "", Range{}, false, nil
}

// Diagnostics returns the diagnostics for the documents that have the given path prefix. This method
// also returns the size of the complete result set to aid in pagination.
func (s *Store) Diagnostics(ctx context.Context, bundleID int, prefix string, limit, offset int) (_ []Diagnostic, _ int, err error) {
	ctx, traceLog, endObservation := s.operations.diagnostics.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("prefix", prefix),
		log.Int("limit", limit),
		log.Int("offset", offset),
	}})
	defer endObservation(1, observation.Args{})

	documentData, err := s.scanDocumentData(s.Store.Query(ctx, sqlf.Sprintf(diagnosticsQuery, bundleID, prefix+"%")))
	if err != nil {
		return nil, 0, err
	}
	traceLog(log.Int("numDocuments", len(documentData)))

	totalCount := 0
	for _, documentData := range documentData {
		totalCount += len(documentData.Document.Diagnostics)
	}
	traceLog(log.Int("totalCount", totalCount))

	diagnostics := make([]Diagnostic, 0, limit)
	for _, documentData := range documentData {
		for _, diagnostic := range documentData.Document.Diagnostics {
			offset--

			if offset < 0 && len(diagnostics) < limit {
				diagnostics = append(diagnostics, Diagnostic{
					DumpID:         bundleID,
					Path:           documentData.Path,
					DiagnosticData: diagnostic,
				})
			}
		}
	}

	return diagnostics, totalCount, nil
}

const diagnosticsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:Diagnostics
SELECT dump_id, path, data FROM lsif_data_documents WHERE dump_id = %s AND path LIKE %s ORDER BY path
`

// MonikersByPosition returns all monikers attached ranges containing the given position. If multiple
// ranges contain the position, then this method will return multiple sets of monikers. Each slice
// of monikers are attached to a single range. The order of the output slice is "outside-in", so that
// the range attached to earlier monikers enclose the range attached to later monikers.
func (s *Store) MonikersByPosition(ctx context.Context, bundleID int, path string, line, character int) (_ [][]MonikerData, err error) {
	ctx, traceLog, endObservation := s.operations.monikersByPosition.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(documentQuery, bundleID, path)))
	if err != nil || !exists {
		return nil, err
	}

	traceLog(log.Int("numRanges", len(documentData.Document.Ranges)))
	ranges := FindRanges(documentData.Document.Ranges, line, character)
	traceLog(log.Int("numIntersectingRanges", len(ranges)))

	monikerData := make([][]MonikerData, 0, len(ranges))
	for _, r := range ranges {
		batch := make([]MonikerData, 0, len(r.MonikerIDs))
		for _, monikerID := range r.MonikerIDs {
			if moniker, exists := documentData.Document.Monikers[monikerID]; exists {
				batch = append(batch, moniker)
			}
		}
		traceLog(log.Int("numMonikersForRange", len(batch)))

		monikerData = append(monikerData, batch)
	}
	traceLog(log.Int("numMonikers", len(monikerData)))

	return monikerData, nil
}

// BulkMonikerResults returns the locations within one of the given bundles that define or reference
// one of the given monikers. This method also returns the size of the complete result set to aid in
// pagination.
func (s *Store) BulkMonikerResults(ctx context.Context, tableName string, uploadIDs []int, monikers []MonikerData, limit, offset int) (_ []Location, _ int, err error) {
	ctx, traceLog, endObservation := s.operations.bulkMonikerResults.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
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

	locationData, err := s.scanQualifiedMonikerLocations(s.Store.Query(ctx, sqlf.Sprintf(
		bulkMonikerResultsQuery,
		sqlf.Sprintf(fmt.Sprintf("lsif_data_%s", tableName)),
		sqlf.Join(idQueries, ", "),
		sqlf.Join(monikerQueries, ", "),
	)))
	if err != nil {
		return nil, 0, err
	}

	totalCount := 0
	for _, monikerLocations := range locationData {
		totalCount += len(monikerLocations.Locations)
	}
	traceLog(
		log.Int("numDumps", len(locationData)),
		log.Int("totalCount", totalCount),
	)

	max := totalCount
	if totalCount > limit {
		max = limit
	}

	locations := make([]Location, 0, max)
outer:
	for _, monikerLocations := range locationData {
		for _, row := range monikerLocations.Locations {
			offset--
			if offset >= 0 {
				continue
			}

			locations = append(locations, Location{
				DumpID: monikerLocations.DumpID,
				Path:   row.URI,
				Range:  newRange(row.StartLine, row.StartCharacter, row.EndLine, row.EndCharacter),
			})

			if len(locations) >= limit {
				break outer
			}
		}
	}
	traceLog(log.Int("numLocations", len(locations)))

	return locations, totalCount, nil
}

const bulkMonikerResultsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:BulkMonikerResults
SELECT dump_id, scheme, identifier, data FROM %s WHERE dump_id IN (%s) AND (scheme, identifier) IN (%s) ORDER BY (scheme, identifier, dump_id)
`

// PackageInformation looks up package information data by identifier.
func (s *Store) PackageInformation(ctx context.Context, bundleID int, path, packageInformationID string) (_ PackageInformationData, _ bool, err error) {
	ctx, endObservation := s.operations.packageInformation.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.String("packageInformationID", packageInformationID),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(packageInformationQuery, bundleID, path)))
	if err != nil || !exists {
		return PackageInformationData{}, false, err
	}

	packageInformationData, exists := documentData.Document.PackageInformation[ID(packageInformationID)]
	return packageInformationData, exists, nil
}

const packageInformationQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:PackageInformation
SELECT dump_id, path, data FROM lsif_data_documents WHERE dump_id = %s AND path = %s LIMIT 1
`

// locations queries the locations associated with the given definition or reference identifiers This
// method returns a map from result set identifiers to another map from document paths to locations
// within that document.
func (s *Store) locations(ctx context.Context, bundleID int, ids []ID) (_ map[ID][]Location, err error) {
	ctx, traceLog, endObservation := s.operations.locations.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.Int("numIDs", len(ids)),
		log.String("ids", idsToString(ids)),
	}})
	defer endObservation(1, observation.Args{})

	if len(ids) == 0 {
		return nil, nil
	}

	// Get the list of indexes we need to read in order to find each result set identifier
	indexes, err := s.translateIDsToResultChunkIndexes(ctx, bundleID, ids)
	if err != nil {
		return nil, err
	}
	traceLog(
		log.Int("numIndexes", len(indexes)),
		log.String("indexes", intsToString(indexes)),
	)

	// Read the result sets and construct the set of documents we need to open to resolve range
	// identifiers into actual offsets in a document.
	paths, rangeIDsByResultID, err := s.readLocationsFromResultChunks(ctx, bundleID, ids, indexes)
	if err != nil {
		return nil, err
	}
	traceLog(
		log.Int("numPaths", len(paths)),
		log.String("paths", strings.Join(paths, ", ")),
	)

	// Hydrate the locations result set by replacing range ids with their actual data from their
	// containing document. This refines the map constructed in the previous step.
	locationsByResultID, totalCount, err := s.readRangesFromDocuments(ctx, bundleID, ids, paths, rangeIDsByResultID, traceLog)
	if err != nil {
		return nil, err
	}
	traceLog(log.Int("numLocations", totalCount))

	return locationsByResultID, nil
}

// ErrNoMetadata occurs if we can't determine the number of result chunks for an index.
var ErrNoMetadata = errors.New("no rows in meta table")

// translateIDsToResultChunkIndexes converts a set of result set identifiers within a given bundle into a
// deduplicated and sorted set of result chunk indexes that compoletely cover those identifiers.
func (s *Store) translateIDsToResultChunkIndexes(ctx context.Context, bundleID int, ids []ID) ([]int, error) {
	// Mapping ids to result chunk indexes relies on the number of total result chunks written during
	// processing so that we can hash identifiers to their parent result chunk in the same deterministic
	// way.
	numResultChunks, exists, err := basestore.ScanFirstInt(s.Store.Query(ctx, sqlf.Sprintf(translateIDsToResultChunkIndexesQuery, bundleID)))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNoMetadata
	}

	resultChunkIndexMap := map[int]struct{}{}
	for _, id := range ids {
		resultChunkIndexMap[HashKey(id, numResultChunks)] = struct{}{}
	}

	indexes := make([]int, 0, len(resultChunkIndexMap))
	for index := range resultChunkIndexMap {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)

	return indexes, nil
}

const translateIDsToResultChunkIndexesQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:translateIDsToResultChunkIndexes
SELECT num_result_chunks FROM lsif_data_metadata WHERE dump_id = %s
`

// resultChunkBatchSize is the maximum number of result chunks we will query at once to resolve a single
// locations request.
const resultChunkBatchSize = 50

// readLocationsFromResultChunks reads the given result chunk indexes for a given bundle. This method returns
// a map from documents to range identifiers that compose each of the given input result set identifiers. This
// method also returns a deduplicated and sorted set of document paths that are referenced in the output map.
func (s *Store) readLocationsFromResultChunks(ctx context.Context, bundleID int, ids []ID, indexes []int) ([]string, map[ID]map[string][]ID, error) {
	pathMap := map[string]struct{}{}
	rangeIDsByResultID := make(map[ID]map[string][]ID, len(ids))

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
		visitResultChunks := s.makeResultChunkVisitor(s.Store.Query(ctx, sqlf.Sprintf(
			readLocationsFromResultChunksQuery,
			bundleID,
			sqlf.Join(indexQueries, ","),
		)))

		if err := visitResultChunks(func(index int, resultChunkData ResultChunkData) {
			for _, id := range ids {
				documentIDRangeIDs, exists := resultChunkData.DocumentIDRangeIDs[id]
				if !exists {
					continue
				}

				rangeIDsByDocument := make(map[string][]ID, len(documentIDRangeIDs))
				for _, documentIDRangeID := range documentIDRangeIDs {
					if path, ok := resultChunkData.DocumentPaths[documentIDRangeID.DocumentID]; ok {
						pathMap[path] = struct{}{}
						rangeIDsByDocument[path] = append(rangeIDsByDocument[path], documentIDRangeID.RangeID)
					}
				}
				rangeIDsByResultID[id] = rangeIDsByDocument
			}
		}); err != nil {
			return nil, nil, err
		}
	}

	paths := make([]string, 0, len(pathMap))
	for path := range pathMap {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	return paths, rangeIDsByResultID, nil
}

const readLocationsFromResultChunksQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:readLocationsFromResultChunks
SELECT idx, data FROM lsif_data_result_chunks WHERE dump_id = %s AND idx IN (%s)
`

// documentBatchSize is the maximum number of documents we will query at once to resolve a single locations request.
const documentBatchSize = 50

// readRangesFromDocuments reads the given documents for a given bundle. This method returns a map from result set
// identifiers to the set of locations composing that result set. The output resolves the missing data given via the
// rangeIDsByResultID parameter. This method also returns a total count of ranges in the result set.
func (s *Store) readRangesFromDocuments(ctx context.Context, bundleID int, ids []ID, paths []string, rangeIDsByResultID map[ID]map[string][]ID, traceLog observation.TraceLogger) (map[ID][]Location, int, error) {
	totalCount := 0
	locationsByResultID := make(map[ID][]Location, len(ids))

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

		pathQueries := make([]*sqlf.Query, 0, len(batch))
		for _, path := range batch {
			pathQueries = append(pathQueries, sqlf.Sprintf("%s", path))
		}
		visitDocuments := s.makeDocumentVisitor(s.Store.Query(ctx, sqlf.Sprintf(readRangesFromDocumentsQuery, bundleID, sqlf.Join(pathQueries, ","))))

		if err := visitDocuments(func(path string, document DocumentData) {
			for id, rangeIDsByPath := range rangeIDsByResultID {
				rangeIDs := rangeIDsByPath[path]
				if len(rangeIDs) == 0 {
					continue
				}

				locations := make([]Location, 0, len(rangeIDs))
				for _, rangeID := range rangeIDs {
					if r, exists := document.Ranges[rangeID]; exists {
						locations = append(locations, Location{
							DumpID: bundleID,
							Path:   path,
							Range:  newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter),
						})
					}
				}
				traceLog(
					log.String("id", string(id)),
					log.String("path", path),
					log.Int("numLocationsForIDInPath", len(locations)),
				)

				totalCount += len(locations)
				locationsByResultID[id] = append(locationsByResultID[id], locations...)
				sortLocations(locationsByResultID[id])
			}
		}); err != nil {
			return nil, 0, err
		}
	}

	return locationsByResultID, totalCount, nil
}

const readRangesFromDocumentsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:readRangesFromDocuments
SELECT path, data FROM lsif_data_documents WHERE dump_id = %s AND path IN (%s)
`

// sortLocationssorts locations by document, then by offset within a document.
func sortLocations(locations []Location) {
	sort.Slice(locations, func(i, j int) bool {
		if locations[i].Path == locations[j].Path {
			return compareBundleRanges(locations[i].Range, locations[j].Range)
		}

		return strings.Compare(locations[i].Path, locations[j].Path) < 0
	})
}

// compareBundleRanges returns true if r1's start position occurs before r2's start position.
func compareBundleRanges(r1, r2 Range) bool {
	cmp := r1.Start.Line - r2.Start.Line
	if cmp == 0 {
		cmp = r1.Start.Character - r2.Start.Character
	}

	return cmp < 0
}

func newRange(startLine, startCharacter, endLine, endCharacter int) Range {
	return Range{
		Start: Position{
			Line:      startLine,
			Character: startCharacter,
		},
		End: Position{
			Line:      endLine,
			Character: endCharacter,
		},
	}
}

// extractResultIDs extracts result identifiers from each range in the given list.
// The output list is relative to the input range list, but with duplicates removed.
func extractResultIDs(ranges []RangeData, fn func(r RangeData) ID) []ID {
	resultIDs := make([]ID, 0, len(ranges))
	resultIDMap := make(map[ID]struct{}, len(ranges))

	for _, r := range ranges {
		resultID := fn(r)

		if _, ok := resultIDMap[resultID]; !ok && resultID != "" {
			resultIDs = append(resultIDs, resultID)
			resultIDMap[resultID] = struct{}{}
		}
	}

	return resultIDs
}

func monikersToString(vs []MonikerData) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, fmt.Sprintf("%s:%s", v.Scheme, v.Identifier))
	}

	return strings.Join(strs, ", ")
}

func idsToString(vs []ID) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, string(v))
	}

	return strings.Join(strs, ", ")
}
