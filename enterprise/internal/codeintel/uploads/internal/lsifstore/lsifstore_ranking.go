package lsifstore

import (
	"bytes"
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	db "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) InsertDefinitionsAndReferencesForRanking(
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
const batchNumber = 10000

func (s *store) InsertDefintionsForRanking(ctx context.Context, defintions []shared.RankingDefintions) (err error) {
	ctx, _, endObservation := s.operations.setDefinitionsForRanking.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	inserter := func(inserter *batch.Inserter) error {
		batchDefinitions := make([]shared.RankingDefintions, 0, batchNumber)
		for _, def := range defintions {
			batchDefinitions = append(batchDefinitions, def)

			if len(batchDefinitions) == batchNumber {
				fmt.Println("inserting def batch")
				if err := insertDefinitions(ctx, inserter, batchDefinitions); err != nil {
					return err
				}
				batchDefinitions = make([]shared.RankingDefintions, 0, batchNumber)
				fmt.Println("finish inserting def batch")
			}
		}

		if len(batchDefinitions) > 0 {
			if err := insertDefinitions(ctx, inserter, batchDefinitions); err != nil {
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
			"document_root",
			"document_path",
		},
		inserter,
	); err != nil {
		return err
	}

	fmt.Println("finish inserting all def batch")

	return nil
}

func insertDefinitions(
	ctx context.Context,
	inserter *batch.Inserter,
	definitions []shared.RankingDefintions,
) error {
	for _, def := range definitions {
		if err := inserter.Insert(
			ctx,
			def.UploadID,
			def.SymbolName,
			def.Repository,
			def.DocumentRoot,
			def.DocumentPath,
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *store) InsertReferencesForRanking(ctx context.Context, references shared.RankingReferences) (err error) {
	ctx, _, endObservation := s.operations.setReferencesForRanking.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	inserter := func(inserter *batch.Inserter) error {
		batchSymbolNames := make([]string, 0, batchNumber)
		for _, ref := range references.SymbolName {
			batchSymbolNames = append(batchSymbolNames, ref)

			if len(batchSymbolNames) == batchNumber {
				fmt.Println("inserting ref batch")

				if err := inserter.Insert(ctx, references.UploadID, pq.Array(batchSymbolNames)); err != nil {
					return err
				}
				batchSymbolNames = make([]string, 0, batchNumber)

				fmt.Println("finish inserting ref batch")
			}
		}

		if len(batchSymbolNames) > 0 {
			if err := inserter.Insert(ctx, references.UploadID, pq.Array(batchSymbolNames)); err != nil {
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
		[]string{"upload_id", "symbol_name"},
		inserter,
	); err != nil {
		return err
	}

	fmt.Println("finish inserting all ref batch")

	return nil
}
