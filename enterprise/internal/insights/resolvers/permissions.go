package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// FetchUnauthorizedRepos returns a list of repo IDs that the current user does *not* have access to. The primary
// purpose of this is to quickly resolve permissions at query time from the primary postgres database and filter
// code insights in the timeseries database. This approach makes the assumption that most users have access to most
// repos - which is highly likely given the public / private model that repos use today.
func FetchUnauthorizedRepos(ctx context.Context, db dbutil.DB) (results []api.RepoID, err error) {
	store := database.Repos(db)
	conds, err := database.AuthzQueryConds(ctx, db)
	if err != nil {
		return []api.RepoID{}, err
	}

	q := sqlf.Join([]*sqlf.Query{sqlf.Sprintf(fetchUnauthorizedReposSql), conds}, " ")

	rows, err := store.Query(ctx, q)
	if err != nil {
		return []api.RepoID{}, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var temp int
		if err := rows.Scan(&temp); err != nil {
			return []api.RepoID{}, err
		}
		results = append(results, api.RepoID(temp))
	}

	return results, nil
}

const fetchUnauthorizedReposSql = `
-- source: enterprise/internal/insights/resolver/permissions.go:FetchUnauthorizedRepos
	SELECT id FROM repo WHERE NOT`
