package lsifstore

import (
	"bytes"
	"context"

	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) ScanDocuments(ctx context.Context, id int, f func(path string, document *scip.Document) error) (err error) {
	ctx, _, endObservation := s.operations.scanDocuments.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	return runQuery(ctx, s.db, sqlf.Sprintf(scanDocumentsQuery, id), func(dbs dbutil.Scanner) error {
		var path string
		var compressedSCIPPayload []byte
		if err := dbs.Scan(&path, &compressedSCIPPayload); err != nil {
			return err
		}

		scipPayload, err := shared.Decompressor.Decompress(bytes.NewReader(compressedSCIPPayload))
		if err != nil {
			return err
		}

		var document scip.Document
		if err := proto.Unmarshal(scipPayload, &document); err != nil {
			return err
		}

		return f(path, &document)
	})
}

const scanDocumentsQuery = `
SELECT
	sid.document_path,
	sd.raw_scip_payload
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE sid.upload_id = %s
ORDER BY sid.document_path
`

func runQuery(ctx context.Context, store *basestore.Store, query *sqlf.Query, f func(dbutil.Scanner) error) (err error) {
	rows, queryErr := store.Query(ctx, query)
	if queryErr != nil {
		return queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := f(rows); err != nil {
			return err
		}
	}

	return nil
}
