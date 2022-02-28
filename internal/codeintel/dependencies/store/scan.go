package store

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type DependencyRepo struct {
	ID      int
	Scheme  string
	Name    string
	Version string
}

func scanDependencyRepos(rows *sql.Rows, queryErr error) (dependencies []DependencyRepo, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var dependencyRepo DependencyRepo

		if err = rows.Scan(
			&dependencyRepo.ID,
			&dependencyRepo.Scheme,
			&dependencyRepo.Name,
			&dependencyRepo.Version,
		); err != nil {
			return nil, err
		}

		dependencies = append(dependencies, dependencyRepo)
	}

	return dependencies, nil
}
