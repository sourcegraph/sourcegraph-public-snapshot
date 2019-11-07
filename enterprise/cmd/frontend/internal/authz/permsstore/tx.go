package permsstore

import (
	"database/sql"
)

// Tx is a transaction object.
type Tx struct {
	*sql.Tx
}

func (tx *Tx) commitOrRollback(err error) {
	if err == nil {
		_ = tx.Commit()
	} else {
		_ = tx.Rollback()
	}
}
