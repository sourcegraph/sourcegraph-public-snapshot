package database

import (
	"context"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type OwnershipStore interface {
	UpdateOwnership(ctx context.Context, repoID api.RepoID, path string, method string, spanByPerson map[string]int) error
}

type ownershipStore struct {
	*basestore.Store
}

func OwnershipsWith(other basestore.ShareableStore) OwnershipStore {
	return &ownershipStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *ownershipStore) UpdateOwnership(ctx context.Context, repoID api.RepoID, path string, method string, spanByPerson map[string]int) error {
	if strings.HasPrefix(path, "/") {
		return errors.New("path cannot start with /")
	}
	for p := path; p != ""; p = dir(p) {
		_, err := s.artifactExists(ctx, repoID, dir(p))
		if err != nil {
			return err
		}
	}
	_, err := s.artifactExists(ctx, repoID, path)
	if err != nil {
		return err
	}
	return nil
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
