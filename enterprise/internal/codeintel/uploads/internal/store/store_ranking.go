package store

import (
	"bytes"
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) InsertDefinitionsAndReferencesForDocument(
	ctx context.Context,
	upload ExportedUpload,
	rankingGraphKey string,
	rankingBatchNumber int,
	setDefsAndRefs func(ctx context.Context, upload ExportedUpload, rankingBatchNumber int, rankingGraphKey, path string, document *scip.Document) error,
) (err error) {
	ctx, _, endObservation := s.operations.insertDefinitionsAndReferencesForDocument.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
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

		scipPayload, err := shared.Decompressor.Decompress(bytes.NewReader(compressedSCIPPayload))
		if err != nil {
			return err
		}

		var document scip.Document
		if err := proto.Unmarshal(scipPayload, &document); err != nil {
			return err
		}
		err = setDefsAndRefs(ctx, upload, rankingBatchNumber, rankingGraphKey, path, &document)
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

func (s *store) InsertDefintionsForRanking(
	ctx context.Context,
	rankingGraphKey string,
	rankingBatchNumber int,
	defintions []shared.RankingDefintions,
) (err error) {
	ctx, _, endObservation := s.operations.insertDefintionsForRanking.With(
		ctx,
		&err,
		observation.Args{},
	)
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	inserter := func(inserter *batch.Inserter) error {
		batchDefinitions := make([]shared.RankingDefintions, 0, rankingBatchNumber)
		for _, def := range defintions {
			batchDefinitions = append(batchDefinitions, def)

			if len(batchDefinitions) == rankingBatchNumber {
				if err := insertDefinitions(ctx, inserter, rankingGraphKey, batchDefinitions); err != nil {
					return err
				}
				batchDefinitions = make([]shared.RankingDefintions, 0, rankingBatchNumber)
			}
		}

		if len(batchDefinitions) > 0 {
			if err := insertDefinitions(ctx, inserter, rankingGraphKey, batchDefinitions); err != nil {
				return err
			}
		}

		return nil
	}

	if err := batch.WithInserter(
		ctx,
		tx.Handle(),
		"codeintel_ranking_definitions",
		batch.MaxNumPostgresParameters,
		[]string{
			"upload_id",
			"symbol_name",
			"repository",
			"document_path",
			"graph_key",
		},
		inserter,
	); err != nil {
		return err
	}

	return nil
}

func insertDefinitions(
	ctx context.Context,
	inserter *batch.Inserter,
	rankingGraphKey string,
	definitions []shared.RankingDefintions,
) error {
	for _, def := range definitions {
		if err := inserter.Insert(
			ctx,
			def.UploadID,
			def.SymbolName,
			def.Repository,
			def.DocumentPath,
			rankingGraphKey,
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *store) InsertReferencesForRanking(
	ctx context.Context,
	rankingGraphKey string,
	rankingBatchNumber int,
	references shared.RankingReferences,
) (err error) {
	ctx, _, endObservation := s.operations.insertReferencesForRanking.With(
		ctx,
		&err,
		observation.Args{},
	)
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	inserter := func(inserter *batch.Inserter) error {
		batchSymbolNames := make([]string, 0, rankingBatchNumber)
		for _, ref := range references.SymbolNames {
			batchSymbolNames = append(batchSymbolNames, ref)

			if len(batchSymbolNames) == rankingBatchNumber {
				if err := inserter.Insert(ctx, references.UploadID, pq.Array(batchSymbolNames), rankingGraphKey); err != nil {
					return err
				}
				batchSymbolNames = make([]string, 0, rankingBatchNumber)
			}
		}

		if len(batchSymbolNames) > 0 {
			if err := inserter.Insert(ctx, references.UploadID, pq.Array(batchSymbolNames), rankingGraphKey); err != nil {
				return err
			}
		}

		return nil
	}

	if err := batch.WithInserter(
		ctx,
		tx.Handle(),
		"codeintel_ranking_references",
		batch.MaxNumPostgresParameters,
		[]string{"upload_id", "symbol_names", "graph_key"},
		inserter,
	); err != nil {
		return err
	}

	return nil
}
