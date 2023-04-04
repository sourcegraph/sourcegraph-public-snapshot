package lsifstore

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetRanges returns definition, reference, implementation, and hover data for each range within the given span of lines.
func (s *store) GetRanges(ctx context.Context, bundleID int, path string, startLine, endLine int) (_ []shared.CodeIntelligenceRange, err error) {
	ctx, _, endObservation := s.operations.getRanges.With(ctx, &err, observation.Args{LogFields: []log.Field{
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
	)))
	if err != nil || !exists {
		return nil, err
	}

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

const rangesDocumentQuery = `
SELECT
	sd.id,
	sid.document_path,
	sd.raw_scip_payload
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE
	sid.upload_id = %s AND
	sid.document_path = %s
LIMIT 1
`

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
