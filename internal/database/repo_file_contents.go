package database

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type RepoFileContentStore interface {
	Create(ctx context.Context, contents string) (int, error)
}

var _ RepoFileContentStore = (*repoFileContentStore)(nil)

// repoFileContentStore handles access to the repo_file_contents table
type repoFileContentStore struct {
	logger log.Logger
	*basestore.Store
}

func RepoFileContentsWith(logger log.Logger, other basestore.ShareableStore) RepoFileContentStore {
	return &repoFileContentStore{
		logger: logger,
		Store:  basestore.NewWithHandle(other.Handle()),
	}
}

// Create
func (s *repoFileContentStore) Create(ctx context.Context, contents string) (int, error) {
	var id int
	if err := s.Handle().QueryRowContext(
		ctx,
		`INSERT INTO repo_file_contents(text_contents)
		VALUES($1)
		RETURNING id`,
		contents,
	).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}
