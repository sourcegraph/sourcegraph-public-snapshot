package lsifstore

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

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

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(locationsDocumentQuery, bundleID, path)))
	if err != nil || !exists {
		return nil, 0, err
	}

	traceLog(log.Int("numRanges", len(documentData.Document.Ranges)))
	ranges := FindRanges(documentData.Document.Ranges, line, character)
	traceLog(log.Int("numIntersectingRanges", len(ranges)))

	orderedResultIDs := extractResultIDs(ranges, extractor)
	locationsMap, totalCount, err := s.locations(ctx, bundleID, orderedResultIDs, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	traceLog(log.Int("totalCount", totalCount))

	locations := make([]Location, 0, limit)
	for _, resultID := range orderedResultIDs {
		locations = append(locations, locationsMap[resultID]...)
	}

	return locations, totalCount, nil
}

// locations queries the locations associated with the given definition or reference identifiers. This
// method returns a map from result set identifiers to another map from document paths to locations
// within that document, as well as a total count of locations within the map.
func (s *Store) locations(ctx context.Context, bundleID int, ids []ID, limit, offset int) (_ map[ID][]Location, _ int, err error) {
	ctx, traceLog, endObservation := s.operations.locations.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
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
	traceLog(
		log.Int("numIndexes", len(indexes)),
		log.String("indexes", intsToString(indexes)),
	)

	// Read the result sets and construct the set of documents we need to open to resolve range
	// identifiers into actual offsets in a document.
	rangeIDsByResultID, totalCount, err := s.readLocationsFromResultChunks(ctx, bundleID, ids, indexes, "")
	if err != nil {
		return nil, 0, err
	}
	traceLog(log.Int("totalCount", totalCount))

	// Filter out all data in rangeIDsByResultID that falls outside of the current page
	rangeIDsByResultID = limitResultMap(ids, rangeIDsByResultID, limit, offset)

	// gather the set of paths for documents we need to fetch from the limited map
	paths := pathsFromResultMap(rangeIDsByResultID)
	traceLog(
		log.Int("numPaths", len(paths)),
		log.String("paths", strings.Join(paths, ", ")),
	)

	// Hydrate the locations result set by replacing range ids with their actual data from their
	// containing document. This refines the map constructed in the previous step.
	locationsByResultID, _, err := s.readRangesFromDocuments(ctx, bundleID, ids, paths, rangeIDsByResultID, traceLog)
	if err != nil {
		return nil, 0, err
	}

	return locationsByResultID, totalCount, nil
}

// limitResultMap returns a map symmetric to the given rangeIDsByResultID including only the location results
// the current page specified by limit and offset.
func limitResultMap(ids []ID, rangeIDsByResultID map[ID]map[string][]ID, limit, offset int) map[ID]map[string][]ID {
	limitedRangeIDsByResultID := make(map[ID]map[string][]ID, len(rangeIDsByResultID))

outer:
	for _, id := range ids {
		rangeIDsByDocument := map[string][]ID{}
		limitedRangeIDsByResultID[id] = rangeIDsByDocument

		paths := make([]string, 0, len(rangeIDsByResultID[id]))
		for path := range rangeIDsByResultID[id] {
			paths = append(paths, path)
		}
		sort.Strings(paths)

		for _, path := range paths {
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

			rangeIDsByDocument[path] = rangeIDs

			if limit == 0 {
				// Hit end of page
				break outer
			}
		}
	}

	return limitedRangeIDsByResultID
}

// pathsFromResultMap returns a deduplicated and sorted set of document paths present in the given map.
func pathsFromResultMap(rangeIDsByResultID map[ID]map[string][]ID) []string {
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

// ErrNoMetadata occurs if we can't determine the number of result chunks for an index.
var ErrNoMetadata = errors.New("no rows in meta table")

// translateIDsToResultChunkIndexes converts a set of result set identifiers within a given bundle into
// a deduplicated and sorted set of result chunk indexes that compoletely cover those identifiers.
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
-- source: enterprise/internal/codeintel/stores/lsifstore/locations.go:translateIDsToResultChunkIndexes
SELECT num_result_chunks FROM lsif_data_metadata WHERE dump_id = %s
`

// resultChunkBatchSize is the maximum number of result chunks we will query at once to resolve a single
// locations request.
const resultChunkBatchSize = 50

// readLocationsFromResultChunks reads the given result chunk indexes for a given bundle. This method returns
// a map from documents to range identifiers that compose each of the given input result set identifiers. If
// a non-empty target path is supplied, then any range falling outside that document path will be omitted from
// the output.
func (s *Store) readLocationsFromResultChunks(ctx context.Context, bundleID int, ids []ID, indexes []int, targetPath string) (map[ID]map[string][]ID, int, error) {
	totalCount := 0
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
-- source: enterprise/internal/codeintel/stores/lsifstore/locations.go:readLocationsFromResultChunks
SELECT idx, data FROM lsif_data_result_chunks WHERE dump_id = %s AND idx IN (%s)
`

// documentBatchSize is the maximum number of documents we will query at once to resolve a single locations request.
const documentBatchSize = 50

// readRangesFromDocuments extracts range data from the documents with the given paths. This method returns a map from
// result set identifiers to the set of locations composing that result set. The output resolves the missing data given
// via the rangeIDsByResultID parameter. This method also returns a total count of ranges in the result set.
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
			totalCount += s.readRangesFromDocument(bundleID, rangeIDsByResultID, locationsByResultID, path, document, traceLog)
		}); err != nil {
			return nil, 0, err
		}
	}

	return locationsByResultID, totalCount, nil
}

const readRangesFromDocumentsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/locations.go:readRangesFromDocuments
SELECT path, data FROM lsif_data_documents WHERE dump_id = %s AND path IN (%s)
`

// readRangesFromDocument extracts range data from the given document. This method populates the given locationsByResultId
// map, which resolves the missing data given via the rangeIDsByResultID parameter. This method returns a total count of
// ranges in the result set.
func (s *Store) readRangesFromDocument(bundleID int, rangeIDsByResultID map[ID]map[string][]ID, locationsByResultID map[ID][]Location, path string, document DocumentData, traceLog observation.TraceLogger) int {
	totalCount := 0
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

	return totalCount
}

// sortLocations sorts locations by document, then by offset within a document.
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

func idsToString(vs []ID) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, string(v))
	}

	return strings.Join(strs, ", ")
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

const locationsDocumentQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/locations.go:{Definitions,References}
SELECT dump_id, path, data FROM lsif_data_documents WHERE dump_id = %s AND path = %s LIMIT 1
`
