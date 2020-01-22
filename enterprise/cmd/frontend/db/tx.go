package db

import (
	"database/sql"
)

// sqlTx is a wrapper over sql.Tx with helper methods.
type sqlTx struct {
	*sql.Tx
}

func (tx *sqlTx) commitOrRollback(err *error) {
	if err == nil || *err == nil {
		_ = tx.Commit()
	} else {
		_ = tx.Rollback()
	}
}
