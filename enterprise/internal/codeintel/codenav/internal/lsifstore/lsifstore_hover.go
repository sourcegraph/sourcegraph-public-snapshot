package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetHover returns the hover text of the symbol at the given position.
func (s *store) GetHover(ctx context.Context, bundleID int, path string, line, character int) (_ string, _ types.Range, _ bool, err error) {
	ctx, trace, endObservation := s.operations.getHover.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
		hoverDocumentQuery,
		bundleID,
		path,
	)))
	if err != nil || !exists {
		return "", types.Range{}, false, err
	}

	if documentData.SCIPData != nil {
		return "", types.Range{}, false, errors.New("SCIP hover unimplemented")
	}

	trace.Log(log.Int("numRanges", len(documentData.LSIFData.Ranges)))
	ranges := precise.FindRanges(documentData.LSIFData.Ranges, line, character)
	trace.Log(log.Int("numIntersectingRanges", len(ranges)))

	for _, r := range ranges {
		if text, ok := documentData.LSIFData.HoverResults[r.HoverResultID]; ok {
			return text, newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter), true, nil
		}
	}

	return "", types.Range{}, false, nil
}

const hoverDocumentQuery = `
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
