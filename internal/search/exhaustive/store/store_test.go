package store_test

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func createUser(db database.DB, username string) (userID int32, err error) {
	admin := username == "admin"

	ctx := context.Background()
	q := sqlf.Sprintf(`INSERT INTO users(username) VALUES(%s, %s) RETURNING id`, username)
	userID, err = basestore.ScanAny[int32](db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...))
	if err != nil {
		return
	}

	roles := []types.SystemRole{types.UserSystemRole}
	if admin {
		roles = append(roles, types.SiteAdministratorSystemRole)
	}

	err = db.UserRoles().BulkAssignSystemRolesToUser(ctx, database.BulkAssignSystemRolesToUserOpts{
		UserID: userID,
		Roles:  roles,
	})
	if err != nil {
		return
	}

	return
}

func createRepo(db database.DB, name string) (api.RepoID, error) {
	repoStore := db.Repos()
	repo := types.Repo{Name: api.RepoName(name)}
	err := repoStore.Create(context.Background(), &repo)
	return repo.ID, err
}
