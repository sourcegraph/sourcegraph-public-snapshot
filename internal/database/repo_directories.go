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
	var id int
	// TODO Add created_at, updated_at
	if err := s.Handle().QueryRowContext(
		ctx,
		`INSERT INTO repo_directories(repo_id, absolute_path, parent_id)
		VALUES($1, $2, $3)
		ON CONFLICT ("repo_id", "absolute_path") DO NOTHING
		RETURNING id`,
		repoID, absolutePath, parentID,
	).Scan(&id); err != nil {
		return nil, err
	}
	pID := 0
	if parentID != nil {
		pID = *parentID
	}
	return &types.RepoDirectory{
		ID:           id,
		RepoID:       repoID,
		AbsolutePath: absolutePath,
		ParentID:     pID,
	}, nil
}
