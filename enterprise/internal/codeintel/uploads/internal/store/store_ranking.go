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
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func (s *store) InsertPathCountInputs(
	ctx context.Context,
	rankingGraphKey string,
	batchSize int,
) (err error) {
	ctx, _, endObservation := s.operations.insertPathCountInputs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if err = s.db.Exec(ctx, sqlf.Sprintf(insertPathCountInputsQuery, rankingGraphKey, batchSize)); err != nil {
		return err
	}

	return nil
}

const insertPathCountInputsQuery = `
WITH
 refs AS (
	SELECT
		id,
		symbol_names
	FROM codeintel_ranking_references rr
	WHERE rr.graph_key = %s AND NOT rr.processed
	ORDER BY rr.symbol_names, rr.id
	FOR UPDATE
	LIMIT %s
),
definitions AS (
	SELECT
		repository,
		document_path,
		graph_key
	FROM codeintel_ranking_definitions
	WHERE symbol_name IN (SELECT unnest(symbol_names) FROM refs)
),
processed AS (
    UPDATE codeintel_ranking_references
    SET processed = true
    WHERE id IN (SELECT r.id FROM refs r)
)
INSERT INTO codeintel_ranking_path_counts_inputs (repository, document_path, count, graph_key)
SELECT
	repository,
	document_path,
	COUNT(*),
	graph_key
FROM definitions
GROUP BY repository, document_path, graph_key
`

func (s *store) InsertPathRanks(
	ctx context.Context,
	graphKey string,
	batchSize int,
) (numPathRanksInserted float64, numInputsProcessed float64, err error) {
	ctx, _, endObservation := s.operations.insertPathRanks.With(
		ctx,
		&err,
		observation.Args{LogFields: []otlog.Field{
			otlog.String("graphKey", graphKey),
		}},
	)
	defer endObservation(1, observation.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(insertPathRanksQuery, graphKey, batchSize))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if !rows.Next() {
		return 0, 0, errors.New("no rows from count")
	}

	if err = rows.Scan(&numPathRanksInserted, &numInputsProcessed); err != nil {
		return 0, 0, err
	}

	return numPathRanksInserted, numInputsProcessed, nil
}

const insertPathRanksQuery = `
WITH input_ranks AS (
    SELECT
        id,
        (SELECT id FROM repo WHERE name = repository) AS repository_id,
        document_path AS path,
        count
    FROM codeintel_ranking_path_counts_inputs
    WHERE graph_key = %s::text AND NOT processed
    ORDER BY graph_key, repository, id
    LIMIT %s
    FOR UPDATE SKIP LOCKED
),
processed AS (
    UPDATE codeintel_ranking_path_counts_inputs
    SET processed = true
    WHERE id IN (SELECT ir.id FROM input_ranks ir)
    RETURNING 1
),
inserted AS (
	INSERT INTO codeintel_path_ranks AS pr (repository_id, precision, payload)
	SELECT
		temp.repository_id,
		1,
		sg_jsonb_concat_agg(temp.row)
	FROM (
		SELECT
			cr.repository_id,
			jsonb_build_object(cr.path, SUM(count)) AS row
		FROM input_ranks cr
		GROUP BY cr.repository_id, cr.path
	) temp
	GROUP BY temp.repository_id
	ON CONFLICT (repository_id, precision) DO UPDATE SET
		graph_key = EXCLUDED.graph_key,
		payload = CASE
			WHEN pr.graph_key != EXCLUDED.graph_key
				THEN EXCLUDED.payload
			ELSE
				(
					SELECT sg_jsonb_concat_agg(row) FROM (
						SELECT jsonb_build_object(key, SUM(value::int)) AS row
						FROM
							(
							SELECT * FROM jsonb_each(pr.payload)
							UNION
							SELECT * FROM jsonb_each(EXCLUDED.payload)
							) AS both_payloads
						GROUP BY key
				) AS combined_json)
			END
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM processed) AS num_processed,
	(SELECT COUNT(*) FROM inserted) AS num_inserted
`
