package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

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
-- source: enterprise/internal/codeintel/stores/lsifstore/diagnostics.go:Diagnostics
SELECT dump_id, path, data FROM lsif_data_documents WHERE dump_id = %s AND path LIKE %s ORDER BY path
`
