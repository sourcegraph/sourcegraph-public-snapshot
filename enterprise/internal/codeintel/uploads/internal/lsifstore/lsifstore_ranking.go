package lsifstore

import (
	"bytes"
	"context"

	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	db "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) CreateDefinitionsAndReferencesForRanking(
	ctx context.Context,
	upload db.ExportedUpload,
	setDefsAndRefs func(ctx context.Context, upload db.ExportedUpload, path string, document *scip.Document) error,
) (err error) {
	ctx, _, endObservation := s.operations.createDefinitionsAndReferencesForRanking.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("id", upload.ID),
	}})
	defer endObservation(1, observation.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(getDocumentsByUploadIDQuery, upload.ID))
	if err != nil {
		return err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var path string
		var compressedSCIPPayload []byte
		if err := rows.Scan(&path, &compressedSCIPPayload); err != nil {
			return err
		}

		scipPayload, err := decompressor.decompress(bytes.NewReader(compressedSCIPPayload))
		if err != nil {
			return err
		}

		var document scip.Document
		if err := proto.Unmarshal(scipPayload, &document); err != nil {
			return err
		}
		err = setDefsAndRefs(ctx, upload, path, &document)
		if err != nil {
			return err
		}
	}

	return nil
}

const getDocumentsByUploadIDQuery = `
SELECT
	sid.document_path,
	sd.raw_scip_payload
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE sid.upload_id = %s
ORDER BY sid.document_path
`
