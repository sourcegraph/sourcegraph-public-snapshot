package pgxerrors

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func IsContraintError(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.ConstraintName == constraint {
			return true
		}
	}
	return false
}
