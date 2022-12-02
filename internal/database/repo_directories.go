package database

import (
	"context"
	"path"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepoDirectoryStore interface {
	CreateIfNotExists(ctx context.Context, repoID api.RepoID, absolutePath string) (*types.RepoDirectory, error)
}

var _ RepoDirectoryStore = (*repoDirectoryStore)(nil)

// repoDirectoryStore handles access to the repo_directories table
type repoDirectoryStore struct {
	logger log.Logger
	*basestore.Store
}

func RepoDirectoryWith(logger log.Logger, other basestore.ShareableStore) RepoDirectoryStore {
	return &repoDirectoryStore{
		logger: logger,
		Store:  basestore.NewWithHandle(other.Handle()),
	}
}

// CreateIfNotExists
func (s *repoDirectoryStore) CreateIfNotExists(ctx context.Context, repoID api.RepoID, absolutePath string) (*types.RepoDirectory, error) {
	if strings.HasPrefix(absolutePath, "/") {
		return nil, errors.New("absolute path does not start with /")
	}
	if strings.HasSuffix(absolutePath, "/") {
		return nil, errors.New("absolute path does not end with /")
	}
	var parentID *int
	// Not root directory, so must have a parent.
	if strings.Contains(absolutePath, "/") {
		row := s.Handle().QueryRowContext(ctx, `
			SELECT id
			FROM repo_directories
			WHERE repo_id = $1 AND absolute_path = $2`,
			repoID, path.Dir(absolutePath),
		)
		// TODO later we will not assume that parent exists in the database.
		if row == nil {
			return nil, errors.New("parent directory does not exist")
		}
		parentID = new(int)
		if err := row.Scan(parentID); err != nil {
			return nil, errors.Wrapf(err, "parent directory ID")
		}
	}
	var d types.RepoDirectory
	// TODO Add created_at, updated_at
	err := s.Handle().QueryRowContext(
		ctx,
		`WITH input_rows(repo_id, absolute_path, parent_id) AS (
			VALUES ($1::integer, $2::text, $3::integer)
		), ins AS (
			INSERT INTO repo_directories(repo_id, absolute_path, parent_id)
			SELECT * FROM input_rows
			ON CONFLICT (repo_id, absolute_path) DO NOTHING
			RETURNING id, repo_id, absolute_path, parent_id
		)
		SELECT id, repo_id, absolute_path, COALESCE(parent_id, 0) FROM ins
		UNION ALL
		SELECT d.id, d.repo_id, d.absolute_path, COALESCE(d.parent_id, 0)
		FROM input_rows
		JOIN repo_directories AS d USING(repo_id, absolute_path)`,
		repoID, absolutePath, parentID,
	).Scan(&d.ID, &d.RepoID, &d.AbsolutePath, &d.ParentID)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
