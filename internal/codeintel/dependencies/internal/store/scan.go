package store

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

func scanDependencyRepos(rows *sql.Rows, queryErr error) (dependencyRepos []shared.Repo, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var dependencyRepo shared.Repo

		if err = rows.Scan(
			&dependencyRepo.ID,
			&dependencyRepo.Scheme,
			&dependencyRepo.Name,
			&dependencyRepo.Version,
		); err != nil {
			return nil, err
		}

		dependencyRepos = append(dependencyRepos, dependencyRepo)
	}

	return dependencyRepos, nil
}
