package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	)))
	if err != nil || !exists {
		return nil, err
	}

	if documentData.SCIPData != nil {
		return nil, errors.New("SCIP stencil unimplemented")
	}

	trace.Log(log.Int("numRanges", len(documentData.LSIFData.Ranges)))

	ranges := make([]types.Range, 0, len(documentData.LSIFData.Ranges))
	for _, r := range documentData.LSIFData.Ranges {
		ranges = append(ranges, newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter))
	}

	return ranges, nil
}

const stencilQuery = `
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
`
