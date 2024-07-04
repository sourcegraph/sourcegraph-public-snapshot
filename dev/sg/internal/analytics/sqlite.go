package analytics

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type retryableConn struct {
	db *sql.DB
}

func (c *retryableConn) Exec(query string, args ...interface{}) (sql.Result, error) {
	for i := 0; i < 2; i++ {
		res, err := c.db.Exec(query, args...)
		if err == nil {
			return res, nil
		}
		var sqliteerr *sqlite.Error
		if errors.As(err, &sqliteerr) && sqliteerr.Code() == sqlite3.SQLITE_BUSY {
			continue
		}
		return nil, err
	}
	return nil, errors.New("sqlite insert failed after multiple attempts due to locking")
}
