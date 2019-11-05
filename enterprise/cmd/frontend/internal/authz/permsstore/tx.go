package permsstore

import (
	"database/sql"
)

type sqlTx struct {
	*sql.Tx
}

func (tx *sqlTx) commitOrRollback(err error) {
	if err == nil {
		_ = tx.Commit()
	} else {
		_ = tx.Rollback()
	}
}
