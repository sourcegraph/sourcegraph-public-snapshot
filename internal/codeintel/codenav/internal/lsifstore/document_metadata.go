package lsifstore

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Stencil returns all ranges within a single document.
func (s *store) GetStencil(ctx context.Context, bundleID int, path core.UploadRelPath) (_ []shared.Range, err error) {
	ctx, trace, endObservation := s.operations.getStencil.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("bundleID", bundleID),
		attribute.String("path", path.RawValue()),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
		stencilQuery,
		bundleID,
		path.RawValue(),
	)))
	if err != nil || !exists {
		return nil, err
	}

	trace.AddEvent("TODO Domain Owner", attribute.Int("numOccurrences", len(documentData.SCIPData.Occurrences)))

	ranges := make([]shared.Range, 0, len(documentData.SCIPData.Occurrences))
	for _, occurrence := range documentData.SCIPData.Occurrences {
		ranges = append(ranges, shared.TranslateRange(scip.NewRangeUnchecked(occurrence.Range)))
	}

	return ranges, nil
}

const stencilQuery = `
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

// GetRanges returns definition, reference, implementation, and hover data for each range within the given span of lines.
func (s *store) GetRanges(ctx context.Context, bundleID int, path core.UploadRelPath, startLine, endLine int) (_ []shared.CodeIntelligenceRange, err error) {
	ctx, _, endObservation := s.operations.getRanges.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("bundleID", bundleID),
		attribute.String("path", path.RawValue()),
		attribute.Int("startLine", startLine),
		attribute.Int("endLine", endLine),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
		rangesDocumentQuery,
		bundleID,
		path.RawValue(),
	)))
	if err != nil || !exists {
		return nil, err
	}

	var ranges []shared.CodeIntelligenceRange
	for _, lookupOccurrence := range documentData.SCIPData.Occurrences {

		r := shared.TranslateRange(scip.NewRangeUnchecked(lookupOccurrence.Range))

		if (startLine <= r.Start.Line && r.Start.Line < endLine) || (startLine <= r.End.Line && r.End.Line < endLine) {
			data := extractOccurrenceData(documentData.SCIPData, lookupOccurrence)

			ranges = append(ranges, shared.CodeIntelligenceRange{
				Range:           r,
				Definitions:     shared.BuildUsages(data.definitions, bundleID, path, shared.UsageKindDefinition),
				References:      shared.BuildUsages(data.references, bundleID, path, shared.UsageKindReference),
				Implementations: shared.BuildUsages(data.implementations, bundleID, path, shared.UsageKindImplementation),
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
