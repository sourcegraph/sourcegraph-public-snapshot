package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Stencil return all ranges within a single document.
func (s *Store) Stencil(ctx context.Context, bundleID int, path string) (_ []Range, err error) {
	ctx, trace, endObservation := s.operations.stencil.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(rangesDocumentQuery, bundleID, path)))
	if err != nil || !exists {
		return nil, err
	}

	trace.Log(log.Int("numRanges", len(documentData.Document.Ranges)))

	ranges := make([]Range, 0, len(documentData.Document.Ranges))
	for _, r := range documentData.Document.Ranges {
		ranges = append(ranges, newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter))
	}

	return ranges, nil
}
