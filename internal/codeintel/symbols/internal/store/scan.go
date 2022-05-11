package store

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

func scanSymbols(rows *sql.Rows, queryErr error) (symbols []shared.Symbol, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var symbol shared.Symbol

		if err = rows.Scan(
			&symbol.Name,
		); err != nil {
			return nil, err
		}

		symbols = append(symbols, symbol)
	}

	return symbols, nil
}
