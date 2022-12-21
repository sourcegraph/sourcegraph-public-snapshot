package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Stencil returns all ranges within a single document.
func (s *store) GetStencil(ctx context.Context, bundleID int, path string) (_ []types.Range, err error) {
	ctx, trace, endObservation := s.operations.getStencil.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
		stencilQuery,
		bundleID,
		path,
		bundleID,
		path,
	)))
	if err != nil || !exists {
		return nil, err
	}

	if documentData.SCIPData != nil {
		trace.AddEvent("TODO Domain Owner", attribute.Int("numOccurrences", len(documentData.SCIPData.Occurrences)))

		ranges := make([]types.Range, 0, len(documentData.SCIPData.Occurrences))
		for _, occurrence := range documentData.SCIPData.Occurrences {
			ranges = append(ranges, translateRange(scip.NewRange(occurrence.Range)))
		}

		return ranges, nil
	}

	trace.AddEvent("TODO Domain Owner", attribute.Int("numRanges", len(documentData.LSIFData.Ranges)))

	ranges := make([]types.Range, 0, len(documentData.LSIFData.Ranges))
	for _, r := range documentData.LSIFData.Ranges {
		ranges = append(ranges, newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter))
	}

	return ranges, nil
}

const stencilQuery = `
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
