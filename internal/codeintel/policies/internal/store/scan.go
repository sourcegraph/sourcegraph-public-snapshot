package store

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

func scanPolicies(rows *sql.Rows, queryErr error) (policies []shared.Policy, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var policy shared.Policy

		if err = rows.Scan(
			&policy.ID,
		); err != nil {
			return nil, err
		}

		policies = append(policies, policy)
	}

	return policies, nil
}
