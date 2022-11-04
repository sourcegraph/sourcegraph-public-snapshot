package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

const documentsLimit = 100

func (s *store) GetUploadDocumentsForPath(ctx context.Context, bundleID int, pathPattern string) ([]string, int, error) {
	totalCount, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(documentsCountQuery, bundleID, pathPattern)))
	if err != nil {
		return nil, 0, err
	}

	documents, err := basestore.ScanStrings(s.db.Query(ctx, sqlf.Sprintf(documentsQuery, bundleID, pathPattern, documentsLimit)))
	if err != nil {
		return nil, 0, err
	}

	return documents, totalCount, err
}

const documentsCountQuery = `
-- source: internal/codeintel/stores/lsifstore/documents.go:Documents
SELECT
	COUNT(*)
FROM
	lsif_data_documents
WHERE
	dump_id = %s AND
	path ILIKE %s
`

const documentsQuery = `
-- source: internal/codeintel/stores/lsifstore/documents.go:Documents
SELECT
	path
FROM
	lsif_data_documents
WHERE
	dump_id = %s AND
	path ILIKE %s
ORDER BY path
LIMIT %s
`
