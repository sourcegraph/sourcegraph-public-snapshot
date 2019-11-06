package permsstore

import (
	"database/sql"
)

// Tx is a transaction object.
type Tx struct {
	*sql.Tx
}

// CommitOrRollback commits changes in the transaction if error is nil,
// otherwise it rolls back.
func (tx *Tx) CommitOrRollback(err error) {
	if err == nil {
		_ = tx.Commit()
	} else {
		_ = tx.Rollback()
	}
}
