package httpapi

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

type DBStore interface {
	Transact(ctx context.Context) (DBStore, error)
	Done(err error) error

	GetUploadByID(ctx context.Context, uploadID int) (dbstore.Upload, bool, error)
	InsertUpload(ctx context.Context, upload dbstore.Upload) (int, error)
	AddUploadPart(ctx context.Context, uploadID, partIndex int) error
	MarkQueued(ctx context.Context, id int, uploadSize *int64) error
	MarkFailed(ctx context.Context, id int, reason string) error
}

type DBStoreShim struct {
	*dbstore.Store
}

func (s *DBStoreShim) Transact(ctx context.Context) (DBStore, error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &DBStoreShim{tx}, nil
}

type GitHubClient interface {
	GetRepository(ctx context.Context, owner string, name string) (*github.Repository, error)
	ListInstallationRepositories(ctx context.Context, page int) ([]*github.Repository, bool, int, error)
}
