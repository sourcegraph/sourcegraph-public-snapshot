package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const documentsLimit = 100

func (s *store) GetUploadDocumentsForPath(ctx context.Context, bundleID int, pathPattern string) (_ []string, _ int, err error) {
	ctx, _, endObservation := s.operations.getUploadDocumentsForPath.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("bundleID", bundleID),
		otlog.String("pathPattern", pathPattern),
	}})
	defer endObservation(1, observation.Args{})

	totalCount, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(
		documentsCountQuery,
		bundleID,
		pathPattern,
	)))
	if err != nil {
		return nil, 0, err
	}

	documents, err := basestore.ScanStrings(s.db.Query(ctx, sqlf.Sprintf(
		documentsQuery,
		bundleID,
		pathPattern,
		documentsLimit,
	)))
	if err != nil {
		return nil, 0, err
	}

	return documents, totalCount, err
}

const documentsCountQuery = `
SELECT COUNT(*)
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE sid.upload_id = %s AND sid.document_path ILIKE %s
`

const documentsQuery = `
SELECT sid.document_path AS path
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE sid.upload_id = %s AND sid.document_path ILIKE %s
ORDER BY path
LIMIT %s
`
