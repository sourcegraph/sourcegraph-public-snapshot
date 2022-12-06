package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Store provides the interface for ranking storage.
type Store interface {
	// Transactions
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	GetStarRank(ctx context.Context, repoName api.RepoName) (float64, error)
	GetRepos(ctx context.Context) ([]api.RepoName, error)
	GetDocumentRanks(ctx context.Context, repoName api.RepoName) (map[string][2]float64, bool, error)
	SetDocumentRanks(ctx context.Context, repoName api.RepoName, precision float64, ranks map[string]float64) error
	HasInputFilename(ctx context.Context, graphKey string, filenames []string) ([]string, error)
	BulkSetDocumentRanks(ctx context.Context, graphKey, filename string, precision float64, ranks map[api.RepoName]map[string]float64) error
	MergeDocumentRanks(ctx context.Context, graphKey string, inputFileBatchSize int) (numRepositoriesUpdated int, numInputsProcessed int, _ error)
	LastUpdatedAt(ctx context.Context, repoIDs []api.RepoID) (map[api.RepoID]time.Time, error)
	UpdatedAfter(ctx context.Context, t time.Time) ([]api.RepoName, error)

	ExportRankPayloadFor(ctx context.Context, repoName api.RepoName) (time.Time, []byte, error)
}

// store manages the ranking store.
type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

// New returns a new ranking store.
func New(observationCtx *observation.Context, db database.DB) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("ranking.store", ""),
		operations: newOperations(observationCtx),
	}
}

func (s *store) Transact(ctx context.Context) (Store, error) {
	return s.transact(ctx)
}

func (s *store) transact(ctx context.Context) (*store, error) {
	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &store{
		logger:     s.logger,
		db:         tx,
		operations: s.operations,
	}, nil
}

func (s *store) Done(err error) error {
	return s.db.Done(err)
}

func (s *store) GetStarRank(ctx context.Context, repoName api.RepoName) (float64, error) {
	rank, _, err := basestore.ScanFirstFloat(s.db.Query(ctx, sqlf.Sprintf(getStarRankQuery, repoName)))
	return rank, err
}

const getStarRankQuery = `
SELECT
	s.rank
FROM (
	SELECT
		name,
		percent_rank() OVER (ORDER BY stars) AS rank
	FROM repo
) s
WHERE s.name = %s
`

func (s *store) GetRepos(ctx context.Context) ([]api.RepoName, error) {
	names, err := basestore.ScanStrings(s.db.Query(ctx, sqlf.Sprintf(getReposQuery)))
	if err != nil {
		return nil, err
	}

	repoNames := make([]api.RepoName, 0, len(names))
	for _, name := range names {
		repoNames = append(repoNames, api.RepoName(name))
	}

	return repoNames, nil
}

const getReposQuery = `
SELECT r.name FROM repo r
WHERE
	r.deleted_at IS NULL AND
	r.blocked IS NULL
ORDER BY r.name
`

func (s *store) GetDocumentRanks(ctx context.Context, repoName api.RepoName) (map[string][2]float64, bool, error) {
	pathRanksWithPrecision := map[string][2]float64{}
	scanner := func(s dbutil.Scanner) error {
		var (
			precision  float64
			serialized string
		)
		if err := s.Scan(&precision, &serialized); err != nil {
			return err
		}

		pathRanks := map[string]float64{}
		if err := json.Unmarshal([]byte(serialized), &pathRanks); err != nil {
			return err
		}

		for path, newRank := range pathRanks {
			if oldRank, ok := pathRanksWithPrecision[path]; ok && oldRank[0] <= precision {
				continue
			}

			pathRanksWithPrecision[path] = [2]float64{precision, newRank}
		}

		return nil
	}

	if err := basestore.NewCallbackScanner(scanner)(s.db.Query(ctx, sqlf.Sprintf(getDocumentRanksQuery, repoName))); err != nil {
		return nil, false, err
	}
	return pathRanksWithPrecision, true, nil
}

const getDocumentRanksQuery = `
SELECT
	precision,
	payload
FROM codeintel_path_ranks pr
JOIN repo r ON r.id = pr.repository_id
WHERE
	r.name = %s AND
	r.deleted_at IS NULL AND
	r.blocked IS NULL
`

func (s *store) SetDocumentRanks(ctx context.Context, repoName api.RepoName, precision float64, ranks map[string]float64) error {
	serialized, err := json.Marshal(ranks)
	if err != nil {
		return err
	}

	return s.db.Exec(ctx, sqlf.Sprintf(setDocumentRanksQuery, repoName, precision, serialized))
}

const setDocumentRanksQuery = `
INSERT INTO codeintel_path_ranks AS pr (repository_id, precision, payload)
VALUES (
	(SELECT id FROM repo WHERE name = %s),
	%s,
	%s
)
ON CONFLICT (repository_id, precision) DO
UPDATE
	SET payload = EXCLUDED.payload
`

func (s *store) HasInputFilename(ctx context.Context, graphKey string, filenames []string) ([]string, error) {
	// Filter the set of filenames that have a representative row in codeintel_path_rank_inputs
	return basestore.ScanStrings(s.db.Query(ctx, sqlf.Sprintf(hasInputFilenameQuery, pq.Array(filenames), graphKey)))
}

// Encourage n index scans on codeintel_path_rank_inputs, as we're hitting a very fast index. Each
// input filename can have tens of thousands of rows, and we'd prefer a semi-join (EXISTS) which only
// cares to find one matching row over a merge join, which would match a large intermediate result
// which then needs to be aggregated away via DISTINCT. This query shape is a bit odd, but helps to
// encourage the optimal behavior.
const hasInputFilenameQuery = `
SELECT s.input_filename FROM unnest(%s::text[]) AS s(input_filename)
WHERE EXISTS (
	SELECT 1
	FROM codeintel_path_rank_inputs pr
	WHERE
		pr.graph_key = %s AND
		pr.input_filename = s.input_filename
)
`

func (s *store) BulkSetDocumentRanks(ctx context.Context, graphKey, filename string, precision float64, ranks map[api.RepoName]map[string]float64) error {
	inserter := batch.NewInserterWithConflict(
		ctx,
		s.db.Handle(),
		"codeintel_path_rank_inputs",
		batch.MaxNumPostgresParameters,
		"ON CONFLICT DO NOTHING",
		"graph_key",
		"input_filename",
		"repository_name",
		"precision",
		"payload",
	)
	for repoName, ranks := range ranks {
		serialized, err := json.Marshal(ranks)
		if err != nil {
			return err
		}

		if err := inserter.Insert(ctx, graphKey, filename, repoName, precision, serialized); err != nil {
			return err
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return err
	}

	return nil
}

func (s *store) MergeDocumentRanks(ctx context.Context, graphKey string, inputFileBatchSize int) (numRepositoriesUpdated int, numInputsProcessed int, err error) {
	rows, err := s.db.Query(ctx, sqlf.Sprintf(mergeDocumentRanksQuery, graphKey, inputFileBatchSize))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if !rows.Next() {
		return 0, 0, errors.New("no rows from count")
	}

	if err = rows.Scan(&numRepositoriesUpdated, &numInputsProcessed); err != nil {
		return 0, 0, err
	}

	return numRepositoriesUpdated, numInputsProcessed, nil
}

const mergeDocumentRanksQuery = `
WITH
locked_candidates AS (
	SELECT
		pr.id,
		pr.graph_key,
		pr.precision,
		pr.input_filename,
		pr.repository_name,
		pr.payload
	FROM codeintel_path_rank_inputs pr
	WHERE pr.graph_key = %s AND NOT pr.processed
	ORDER BY pr.repository_name, pr.id
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
upserted AS (
	INSERT INTO codeintel_path_ranks AS pr (repository_id, precision, graph_key, payload)
	SELECT
		r.id,
		c.precision,
		c.graph_key,
		sg_jsonb_concat_agg(c.payload)
	FROM locked_candidates c
	JOIN repo r ON r.name = c.repository_name
	GROUP BY r.id, c.precision, c.graph_key
	ON CONFLICT (repository_id, precision) DO UPDATE SET
		graph_key = EXCLUDED.graph_key,
		payload   = CASE
			WHEN pr.graph_key != EXCLUDED.graph_key
				THEN EXCLUDED.payload
			ELSE
				pr.payload || EXCLUDED.payload
		END
	RETURNING 1
),
processed AS (
	UPDATE codeintel_path_rank_inputs
	SET processed = true
	WHERE id IN (SELECT c.id FROM locked_candidates c)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM upserted) AS num_updated,
	(SELECT COUNT(*) FROM processed) AS num_processed
`

func (s *store) LastUpdatedAt(ctx context.Context, repoIDs []api.RepoID) (map[api.RepoID]time.Time, error) {
	pairs, err := scanLastUpdatedAtPairs(s.db.Query(ctx, sqlf.Sprintf(lastUpdatedAtQuery, pq.Array(repoIDs))))
	if err != nil {
		return nil, err
	}

	return pairs, nil
}

const lastUpdatedAtQuery = `
SELECT
	repository_id,
	updated_at
FROM codeintel_path_ranks
WHERE repository_id = ANY(%s)
`

var scanLastUpdatedAtPairs = basestore.NewMapScanner(func(s dbutil.Scanner) (repoID api.RepoID, t time.Time, _ error) {
	err := s.Scan(&repoID, &t)
	return repoID, t, err
})

func (s *store) UpdatedAfter(ctx context.Context, t time.Time) ([]api.RepoName, error) {
	names, err := basestore.ScanStrings(s.db.Query(ctx, sqlf.Sprintf(updatedAfterQuery, t)))
	if err != nil {
		return nil, err
	}

	repoNames := make([]api.RepoName, 0, len(names))
	for _, name := range names {
		repoNames = append(repoNames, api.RepoName(name))
	}

	return repoNames, nil
}

const updatedAfterQuery = `
SELECT r.name
FROM codeintel_path_ranks pr
JOIN repo r ON r.id = pr.repository_id
WHERE pr.updated_at >= %s
ORDER BY r.name
`

func (s *store) ExportRankPayloadFor(ctx context.Context, repoName api.RepoName) (_ time.Time, _ []byte, err error) {
	type st struct {
		lastUpdated time.Time
		payload     []byte
	}
	scan := basestore.NewFirstScanner(func(s dbutil.Scanner) (st, error) {
		var sx st
		err := s.Scan(&sx.lastUpdated, &sx.payload)
		return sx, err
	})

	const targetPrecision = 1
	v, _, err := scan(s.db.Query(ctx, sqlf.Sprintf(exportRankPayloadForQuery, repoName, targetPrecision)))
	if err != nil {
		return time.Time{}, nil, err
	}

	return v.lastUpdated, v.payload, nil
}

const exportRankPayloadForQuery = `
SELECT
	pr.updated_at,
	pr.payload
FROM codeintel_path_ranks pr
JOIN repo r ON r.id = pr.repository_id
WHERE
	r.name = %s AND
	pr.precision = %s AND
	r.deleted_at IS NULL AND
	r.blocked IS NULL
ORDER BY pr.updated_at DESC
LIMIT 1
`
