package store

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type DependencyRepo struct {
	ID      int
	Name    string
	Version string
}

func scanDependencyRepos(rows *sql.Rows, queryErr error) (dependencies []DependencyRepo, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var dep DependencyRepo
		if err = rows.Scan(&dep.ID, &dep.Name, &dep.Version); err != nil {
			return nil, err
		}

		dependencies = append(dependencies, dep)
	}

	return dependencies, nil
}
