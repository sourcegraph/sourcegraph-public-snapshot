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
			fmt.Println("last one inserting def batch")
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

func (s *store) InsertPathCountInputs(ctx context.Context, uploadID int) (err error) {
	ctx, _, endObservation := s.operations.insertPathCountInputs.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("uploadID", uploadID),
	}})
	defer endObservation(1, observation.Args{})

	if err = s.db.Exec(ctx, sqlf.Sprintf(insertPathCountInputsQuery, uploadID)); err != nil {
		return err
	}

	return nil
}

const insertPathCountInputsQuery = `
WITH refs AS (
	SELECT
		unnest(symbol_names) AS symbol_names
	FROM codeintel_ranking_references
	WHERE upload_id = %s
),
definitions AS (
	SELECT
		repository,
		document_root,
		document_path
	FROM codeintel_ranking_definitions
	WHERE symbol_name IN (SELECT symbol_names FROM refs)
)
INSERT INTO codeintel_ranking_path_counts_inputs (repository, document_root, document_path, count, graph_key)
SELECT repository, document_root, document_path, COUNT(*), 'dev'::text FROM definitions GROUP BY repository, document_root, document_path
`

func (s *store) InsertPathRanks(ctx context.Context, graphKey string, batchSize int) (err error) {
	ctx, _, endObservation := s.operations.insertPathRanks.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("graphKey", graphKey),
	}})
	defer endObservation(1, observation.Args{})

	if err := s.db.Exec(ctx, sqlf.Sprintf(`TRUNCATE TABLE codeintel_path_ranks`)); err != nil {
		return err
	}

	if err = s.db.Exec(ctx, sqlf.Sprintf(insertPathRanksQuery)); err != nil {
		return err
	}

	return nil
}

const insertPathRanksQuery = `
WITH
all_current_ranks AS (
    SELECT
        pr.repository_id AS repository_id,
        data.key AS path,
        data.value::text::int AS count
    FROM codeintel_path_ranks pr,
    json_each(pr.payload::json) AS data
),

input_ranks AS (
    SELECT
        (SELECT id FROM repo WHERE name = repository) AS repository_id,
        document_path AS path,
        SUM(count)::int AS count
    FROM codeintel_ranking_path_counts_inputs
    GROUP BY repository, document_path
),

combined_ranks AS (
    SELECT * FROM all_current_ranks
    UNION
    SELECT * FROM input_ranks
)

INSERT INTO codeintel_path_ranks (repository_id, precision, payload)
SELECT
    temp.repository_id,
    1,
    sg_jsonb_concat_agg(temp.row)
FROM (
    SELECT
        cr.repository_id,
        jsonb_build_object(cr.path, SUM(count)) AS row
    FROM combined_ranks cr
    GROUP BY cr.repository_id, cr.path
) temp
GROUP BY temp.repository_id
ON CONFLICT (repository_id, precision) DO UPDATE SET
    payload = EXCLUDED.payload
`

func (s *store) GetRankingDefinitionsBySymbolNames(ctx context.Context, symbolNames []string) (definitions []shared.RankingDefintions, err error) {
	ctx, _, endObservation := s.operations.getRankingDefinitionsBySymbolNames.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(getRankingDefinitionsBySymbolNamesQuery, pq.Array(symbolNames)))
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	definitions = make([]shared.RankingDefintions, 0)
	for rows.Next() {
		var uploadID int
		var symbolName string
		var repository string
		var documentRoot string
		var documentPath string
		if err := rows.Scan(&uploadID, &symbolName, &repository, &documentRoot, &documentPath); err != nil {
			return nil, err
		}

		definitions = append(definitions, shared.RankingDefintions{
			UploadID:     uploadID,
			SymbolName:   symbolName,
			Repository:   repository,
			DocumentRoot: documentRoot,
			DocumentPath: documentPath,
		})
	}

	return definitions, nil
}

const getRankingDefinitionsBySymbolNamesQuery = `
SELECT
	upload_id,
	symbol_name
	repository,
	document_root,
	document_path
FROM codeintel_ranking_definitions
WHERE symbol_name = ANY(%s)
`

func (s *store) GetRankingReferencesByUploadID(ctx context.Context, uploadID int, limit, offset int) (references []shared.RankingReferences, err error) {
	ctx, _, endObservation := s.operations.getRankingReferencesByUploadID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(getRankingReferencesByUploadIDQuery, uploadID, limit, offset))
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	references = make([]shared.RankingReferences, 0)
	for rows.Next() {
		var uploadID int
		var symbolName []string
		if err := rows.Scan(&uploadID, pq.Array(&symbolName)); err != nil {
			return nil, err
		}

		references = append(references, shared.RankingReferences{
			UploadID:   uploadID,
			SymbolName: symbolName,
		})
	}

	return references, nil
}

const getRankingReferencesByUploadIDQuery = `
SELECT
	upload_id,
	symbol_names
FROM codeintel_ranking_references
WHERE upload_id = %s
ORDER BY upload_id
LIMIT %s
OFFSET %s
`

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
			fmt.Println("last one inserting ref batch")
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
		[]string{"upload_id", "symbol_names"},
		inserter,
	); err != nil {
		return err
	}

	fmt.Println("finish inserting all ref batch")

	return nil
}
