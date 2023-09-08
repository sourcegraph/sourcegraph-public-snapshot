package store_test

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func createUser(store *basestore.Store, username string) (int32, error) {
	admin := username == "admin"
	q := sqlf.Sprintf(`INSERT INTO users(username, site_admin) VALUES(%s, %s) RETURNING id`, username, admin)
	return basestore.ScanAny[int32](store.QueryRow(context.Background(), q))
}

func createRepo(db database.DB, name string) (api.RepoID, error) {
	repoStore := db.Repos()
	repo := types.Repo{Name: api.RepoName(name)}
	err := repoStore.Create(context.Background(), &repo)
	return repo.ID, err
}
