package search_test

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/mock"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func createUser(store *basestore.Store, username string) (int32, error) {
	q := sqlf.Sprintf(`INSERT INTO users(username) VALUES(%s) RETURNING id`, username)
	row := store.QueryRow(context.Background(), q)
	var userID int32
	err := row.Scan(&userID)
	return userID, err
}

func cleanupUsers(store *basestore.Store) error {
	q := sqlf.Sprintf(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	return store.Exec(context.Background(), q)
}

func createRepo(db database.DB, name string) (api.RepoID, error) {
	repoStore := db.Repos()
	repo := types.Repo{Name: api.RepoName(name)}
	err := repoStore.Create(context.Background(), &repo)
	return repo.ID, err
}

func cleanupRepos(store *basestore.Store) error {
	q := sqlf.Sprintf(`TRUNCATE TABLE repo`)
	return store.Exec(context.Background(), q)
}

func cleanupSearchJobs(store *basestore.Store) error {
	q := sqlf.Sprintf(`TRUNCATE TABLE exhaustive_search_jobs`)
	return store.Exec(context.Background(), q)
}

type mockSearcher struct {
	mock.Mock
}

var _ service.NewSearcher = &mockSearcher{}

func (m *mockSearcher) NewSearch(ctx context.Context, q string) (service.SearchQuery, error) {
	args := m.Called(ctx, q)
	return args.Get(0).(service.SearchQuery), args.Error(1)
}

type mockSearchQuery struct {
	mock.Mock
}

var _ service.SearchQuery = &mockSearchQuery{}

func (m *mockSearchQuery) RepositoryRevSpecs(ctx context.Context) ([]service.RepositoryRevSpec, error) {
	args := m.Called(ctx)
	return args.Get(0).([]service.RepositoryRevSpec), args.Error(1)
}

func (m *mockSearchQuery) ResolveRepositoryRevSpec(ctx context.Context, spec service.RepositoryRevSpec) ([]service.RepositoryRevision, error) {
	args := m.Called(ctx, spec)
	return args.Get(0).([]service.RepositoryRevision), args.Error(1)
}

func (m *mockSearchQuery) Search(ctx context.Context, revision service.RepositoryRevision, writer service.CSVWriter) error {
	args := m.Called(ctx, revision, writer)
	return args.Error(0)
}
