package store

import (
	"context"
	"encoding/json"

	"github.com/keegancsmith/sqlf"
	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Store provides the interface for ranking storage.
type Store interface {
	// Transactions
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	GetStarRank(ctx context.Context, repoName api.RepoName) (float64, error)
	GetRepos(ctx context.Context) ([]api.RepoName, error)
	GetDocumentRanks(ctx context.Context, repoName api.RepoName) (map[string][]float64, bool, error)
	SetDocumentRanks(ctx context.Context, repoName api.RepoName, ranks map[string][]float64) error
}

// store manages the ranking store.
type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

// New returns a new ranking store.
func New(db database.DB, observationContext *observation.Context) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("ranking.store", ""),
		operations: newOperations(observationContext),
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

func (s *store) GetDocumentRanks(ctx context.Context, repoName api.RepoName) (map[string][]float64, bool, error) {
	serialized, ok, err := basestore.ScanFirstString(s.db.Query(ctx, sqlf.Sprintf(getDocumentRanksQuery, repoName)))
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}

	m := map[string][]float64{}
	err = json.Unmarshal([]byte(serialized), &m)
	return m, true, err
}

const getDocumentRanksQuery = `
SELECT payload
FROM codeintel_path_ranks pr
JOIN repo r ON r.id = pr.repository_id
WHERE
	r.name = %s AND
	r.deleted_at IS NULL AND
	r.blocked IS NULL
`

func (s *store) SetDocumentRanks(ctx context.Context, repoName api.RepoName, ranks map[string][]float64) error {
	serialized, err := json.Marshal(ranks)
	if err != nil {
		return err
	}
	payload := string(serialized)

	return s.db.Exec(ctx, sqlf.Sprintf(setDocumentRanksQuery, repoName, payload))
}

const setDocumentRanksQuery = `
INSERT INTO codeintel_path_ranks AS pr (repository_id, payload)
VALUES (
	(SELECT id FROM repo WHERE name = %s),
	%s
)
ON CONFLICT (repository_id) DO
UPDATE
	SET payload = (pr.payload::jsonb || EXCLUDED.payload::jsonb)::text
`
