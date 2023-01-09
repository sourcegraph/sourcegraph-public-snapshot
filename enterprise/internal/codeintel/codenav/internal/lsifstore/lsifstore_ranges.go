package lsifstore

import (
	"context"
	"sort"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// MaximumRangesDefinitionLocations is the maximum limit when querying definition locations for a
// Ranges request.
const MaximumRangesDefinitionLocations = 10000

// GetRanges returns definition, reference, implementation, and hover data for each range within the given span of lines.
func (s *store) GetRanges(ctx context.Context, bundleID int, path string, startLine, endLine int) (_ []shared.CodeIntelligenceRange, err error) {
	ctx, trace, endObservation := s.operations.getRanges.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("startLine", startLine),
		log.Int("endLine", endLine),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
		rangesDocumentQuery,
		bundleID,
		path,
		bundleID,
		path,
	)))
	if err != nil || !exists {
		return nil, err
	}

	if documentData.SCIPData != nil {
		var ranges []shared.CodeIntelligenceRange
		for _, occurrence := range documentData.SCIPData.Occurrences {
			r := translateRange(scip.NewRange(occurrence.Range))

			if (startLine <= r.Start.Line && r.Start.Line < endLine) || (startLine <= r.End.Line && r.End.Line < endLine) {
				data := extractOccurrenceData(documentData.SCIPData, occurrence)

				ranges = append(ranges, shared.CodeIntelligenceRange{
					Range:           r,
					Definitions:     convertSCIPRangesToLocations(data.definitions, bundleID, path),
					References:      convertSCIPRangesToLocations(data.references, bundleID, path),
					Implementations: convertSCIPRangesToLocations(data.implementations, bundleID, path),
					HoverText:       strings.Join(data.hoverText, "\n"),
				})
			}
		}

		return ranges, nil
	}

	trace.AddEvent("TODO Domain Owner", attribute.Int("numRanges", len(documentData.LSIFData.Ranges)))
	ranges := precise.FindRangesInWindow(documentData.LSIFData.Ranges, startLine, endLine)
	trace.AddEvent("TODO Domain Owner", attribute.Int("numIntersectingRanges", len(ranges)))

	definitionResultIDs := extractResultIDs(ranges, func(r precise.RangeData) precise.ID { return r.DefinitionResultID })
	definitionLocations, _, err := s.locations(ctx, bundleID, definitionResultIDs, MaximumRangesDefinitionLocations, 0)
	if err != nil {
		return nil, err
	}

	referenceResultIDs := extractResultIDs(ranges, func(r precise.RangeData) precise.ID { return r.ReferenceResultID })
	referenceLocations, err := s.getLocationsWithinFile(ctx, bundleID, referenceResultIDs, path, *documentData.LSIFData)
	if err != nil {
		return nil, err
	}

	implementationResultIDs := extractResultIDs(ranges, func(r precise.RangeData) precise.ID { return r.ImplementationResultID })
	implementationLocations, err := s.getLocationsWithinFile(ctx, bundleID, implementationResultIDs, path, *documentData.LSIFData)
	if err != nil {
		return nil, err
	}

	codeintelRanges := make([]shared.CodeIntelligenceRange, 0, len(ranges))
	for _, r := range ranges {
		codeintelRanges = append(codeintelRanges, shared.CodeIntelligenceRange{
			Range:           newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter),
			Definitions:     definitionLocations[r.DefinitionResultID],
			References:      referenceLocations[r.ReferenceResultID],
			Implementations: implementationLocations[r.ImplementationResultID],
			HoverText:       documentData.LSIFData.HoverResults[r.HoverResultID],
		})
	}
	sort.Slice(codeintelRanges, func(i, j int) bool {
		return compareBundleRanges(codeintelRanges[i].Range, codeintelRanges[j].Range)
	})

	return codeintelRanges, nil
}

const rangesDocumentQuery = `
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
		hovers,
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

// getLocationsWithinFile queries the file-local locations associated with the given definition or reference
// identifiers. Like locations, this method returns a map from result set identifiers to another map from
// document paths to locations within that document.
func (s *store) getLocationsWithinFile(ctx context.Context, bundleID int, ids []precise.ID, path string, documentData precise.DocumentData) (_ map[precise.ID][]shared.Location, err error) {
	ctx, trace, endObservation := s.operations.getLocationsWithinFile.With(ctx, &err, observation.Args{LogFields: []log.Field{
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
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("numIndexes", len(indexes)),
		attribute.String("indexes", intsToString(indexes)))

	// Read the result sets and gather the set of range identifiers we need to resolve with
	// the given document data.
	rangeIDsByResultID, _, err := s.readLocationsFromResultChunks(ctx, bundleID, ids, indexes, path)
	if err != nil {
		return nil, err
	}

	// Hydrate the locations result set by replacing range ids with their actual data from their
	// containing document. This refines the map constructed in the previous step.
	locationsByResultID := make(map[precise.ID][]shared.Location, len(ids))
	totalCount := s.readRangesFromDocument(bundleID, rangeIDsByResultID, locationsByResultID, path, documentData, trace)
	trace.AddEvent("TODO Domain Owner", attribute.Int("numLocations", totalCount))

	return locationsByResultID, nil
}

func convertSCIPRangesToLocations(ranges []*scip.Range, dumpID int, path string) []shared.Location {
	locations := make([]shared.Location, 0, len(ranges))
	for _, r := range ranges {
		locations = append(locations, shared.Location{
			DumpID: dumpID,
			Path:   path,
			Range:  translateRange(r),
		})
	}

	return locations
}
