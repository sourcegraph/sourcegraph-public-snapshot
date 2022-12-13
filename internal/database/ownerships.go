package database

import (
	"context"
	"database/sql"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type OwnershipStore interface {
	UpdateOwnership(ctx context.Context, repoID api.RepoID, path string, method string, spanByPerson map[string]int) error
	FetchOwnership(ctx context.Context, repoID api.RepoID, path string) (Ownership, error)
	Transact(ctx context.Context) (OwnershipStore, error)
	Done(error) error
}

type ownershipStore struct {
	*basestore.Store
}

func OwnershipsWith(other basestore.ShareableStore) OwnershipStore {
	return &ownershipStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *ownershipStore) Transact(ctx context.Context) (OwnershipStore, error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &ownershipStore{Store: tx}, nil
}

func (s *ownershipStore) Done(err error) error {
	return s.Store.Done(err)
}

func (s *ownershipStore) UpdateOwnership(ctx context.Context, repoID api.RepoID, path string, method string, importanceByPerson map[string]int) (err error) {
	if strings.HasPrefix(path, "/") {
		return errors.New("path cannot start with /")
	}
	for p := path; p != ""; p = dir(p) {
		_, err := s.artifactExists(ctx, repoID, dir(p))
		if err != nil {
			return err
		}
	}
	id, err := s.artifactExists(ctx, repoID, path)
	if err != nil {
		return err
	}
	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()
	for person, importance := range importanceByPerson {
		q := `
			INSERT INTO own_signals(artifact_id, who, method, importance_indicator)
			VALUES ($1, $2, $3, $4)
		`
		if err := s.Handle().QueryRowContext(ctx, q, id, person, method, importance).Err(); err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

// Ownership for now is just a map from person name to weight.
type Ownership map[string]int

func (s *ownershipStore) FetchOwnership(ctx context.Context, repoID api.RepoID, path string) (Ownership, error) {
	var p *string
	if path != "" {
		p = &path
	}
	q := `
		SELECT s.who, s.importance_indicator
		FROM own_artifacts AS a
		INNER JOIN own_signals AS s
		ON a.id = s.artifact_id
		WHERE a.repo_id = $1
		AND a.absolute_path = $2
	`
	rs, err := s.Handle().QueryContext(ctx, q, repoID, p)
	if err != nil {
		return nil, err
	}
	o := Ownership{}
	for rs.Next() {
		var who string
		var i int
		if err := rs.Scan(&who, &i); err != nil {
			return nil, err
		}
		o[who] = i
	}
	return o, nil
}

func (s *ownershipStore) artifactExists(ctx context.Context, repoID api.RepoID, path string) (int, error) {
	var p *string
	if path != "" {
		p = &path
	}
	q := `
		WITH input_rows(repo_id, absolute_path) AS (
			VALUES($1::integer, $2::text)
		), ins AS (
			INSERT INTO own_artifacts(repo_id, absolute_path)
			SELECT * FROM input_rows
			ON CONFLICT (repo_id, absolute_path) DO NOTHING
			RETURNING id, repo_id, absolute_path
		)
		SELECT id
		FROM ins
		UNION ALL
		SELECT a.id
		FROM input_rows
		JOIN own_artifacts AS a
		USING (repo_id, absolute_path)`
	var id int
	if err := s.Handle().QueryRowContext(ctx, q, repoID, p).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func dir(p string) string {
	if strings.Contains(p, "/") {
		return path.Dir(p)
	}
	return ""
}
