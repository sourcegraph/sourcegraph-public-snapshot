package lsifstore

import (
	"context"
	"sort"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// MaximumRangesDefinitionLocations is the maximum limit when querying definition locations for a
// Ranges request.
const MaximumRangesDefinitionLocations = 10000

// Ranges returns definition, reference, implementation, and hover data for each range within the given span of lines.
func (s *Store) Ranges(ctx context.Context, bundleID int, path string, startLine, endLine int) (_ []CodeIntelligenceRange, err error) {
	ctx, trace, endObservation := s.operations.ranges.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("startLine", startLine),
		log.Int("endLine", endLine),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(rangesDocumentQuery, bundleID, path)))
	if err != nil || !exists {
		return nil, err
	}

	trace.Log(log.Int("numRanges", len(documentData.Document.Ranges)))
	ranges := precise.FindRangesInWindow(documentData.Document.Ranges, startLine, endLine)
	trace.Log(log.Int("numIntersectingRanges", len(ranges)))

	definitionResultIDs := extractResultIDs(ranges, func(r precise.RangeData) precise.ID { return r.DefinitionResultID })
	definitionLocations, _, err := s.locations(ctx, bundleID, definitionResultIDs, MaximumRangesDefinitionLocations, 0)
	if err != nil {
		return nil, err
	}

	referenceResultIDs := extractResultIDs(ranges, func(r precise.RangeData) precise.ID { return r.ReferenceResultID })
	referenceLocations, err := s.locationsWithinFile(ctx, bundleID, referenceResultIDs, path, documentData.Document)
	if err != nil {
		return nil, err
	}

	implementationResultIDs := extractResultIDs(ranges, func(r precise.RangeData) precise.ID { return r.ImplementationResultID })
	implementationLocations, err := s.locationsWithinFile(ctx, bundleID, implementationResultIDs, path, documentData.Document)
	if err != nil {
		return nil, err
	}

	codeintelRanges := make([]CodeIntelligenceRange, 0, len(ranges))
	for _, r := range ranges {
		codeintelRanges = append(codeintelRanges, CodeIntelligenceRange{
			Range:           newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter),
			Definitions:     definitionLocations[r.DefinitionResultID],
			References:      referenceLocations[r.ReferenceResultID],
			Implementations: implementationLocations[r.ImplementationResultID],
			HoverText:       documentData.Document.HoverResults[r.HoverResultID],
		})
	}
	sort.Slice(codeintelRanges, func(i, j int) bool {
		return compareBundleRanges(codeintelRanges[i].Range, codeintelRanges[j].Range)
	})

	return codeintelRanges, nil
}

const rangesDocumentQuery = `
-- source: internal/codeintel/stores/lsifstore/ranges.go:Ranges
SELECT
	dump_id,
	path,
	data,
	ranges,
	hovers,
	NULL AS monikers,
	NULL AS packages,
	NULL AS diagnostics
FROM
	lsif_data_documents
WHERE
	dump_id = %s AND
	path = %s
LIMIT 1
`

// locationsWithinFile queries the file-local locations associated with the given definition or reference
// identifiers. Like locations, this method returns a map from result set identifiers to another map from
// document paths to locations within that document.
func (s *Store) locationsWithinFile(ctx context.Context, bundleID int, ids []precise.ID, path string, documentData precise.DocumentData) (_ map[precise.ID][]Location, err error) {
	ctx, trace, endObservation := s.operations.locationsWithinFile.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.Int("numIDs", len(ids)),
		log.String("ids", idsToString(ids)),
		log.String("path", path),
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
	trace.Log(
		log.Int("numIndexes", len(indexes)),
		log.String("indexes", intsToString(indexes)),
	)

	// Read the result sets and gather the set of range identifiers we need to resolve with
	// the given document data.
	rangeIDsByResultID, _, err := s.readLocationsFromResultChunks(ctx, bundleID, ids, indexes, path)
	if err != nil {
		return nil, err
	}

	// Hydrate the locations result set by replacing range ids with their actual data from their
	// containing document. This refines the map constructed in the previous step.
	locationsByResultID := make(map[precise.ID][]Location, len(ids))
	totalCount := s.readRangesFromDocument(bundleID, rangeIDsByResultID, locationsByResultID, path, documentData, trace)
	trace.Log(log.Int("numLocations", totalCount))

	return locationsByResultID, nil
}
