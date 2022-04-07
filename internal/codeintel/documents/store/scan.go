package store

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type Todo struct {
	ID      int
	Version string
}

func scanTodos(rows *sql.Rows, queryErr error) (todos []Todo, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var todo Todo

		if err = rows.Scan(
			&todo.ID,
			&todo.Version,
		); err != nil {
			return nil, err
		}

		todos = append(todos, todo)
	}

	return todos, nil
}
