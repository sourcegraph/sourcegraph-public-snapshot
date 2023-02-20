package database

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type CodeownersStore interface {
	basestore.ShareableStore
	Done(error) error

	CreateCodeownersForRepo(ctx context.Context, codeowners *types.CodeownersFile, id api.RepoID) error
	GetCodeownersForRepo(ctx context.Context, id api.RepoID) (*types.CodeownersFile, error)
	DeleteCodeownersForRepo(ctx context.Context, id api.RepoID) error
	ListCodeowners(ctx context.Context) ([]*types.CodeownersFile, error)
}

type codeownersStore struct {
	*basestore.Store
}

func CodeownersWith(other basestore.ShareableStore) CodeownersStore {
	return &codeownersStore{
		Store: basestore.NewWithHandle(other.Handle()),
	}
}

func (s *codeownersStore) With(other basestore.ShareableStore) CodeownersStore {
	return &codeownersStore{
		Store: s.Store.With(other),
	}
}

func (s *codeownersStore) WithTransact(ctx context.Context, f func(store CodeownersStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&codeownersStore{
			Store: tx,
		})
	})
}

func (s *codeownersStore) CreateCodeownersForRepo(ctx context.Context, codeowners *types.CodeownersFile, id api.RepoID) error {
	//TODO implement me
	panic("implement me")
}

func (s *codeownersStore) GetCodeownersForRepo(ctx context.Context, id api.RepoID) (*types.CodeownersFile, error) {
	//TODO implement me
	panic("implement me")
}

func (s *codeownersStore) DeleteCodeownersForRepo(ctx context.Context, id api.RepoID) error {
	//TODO implement me
	panic("implement me")
}

func (s *codeownersStore) ListCodeowners(ctx context.Context) ([]*types.CodeownersFile, error) {
	//TODO implement me
	panic("implement me")
}
