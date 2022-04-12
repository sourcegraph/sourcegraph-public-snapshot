package database

import (
	"database/sql"

	"github.com/grafana/regexp"
	"github.com/mattn/go-sqlite3"
)

func Init() {
	sql.Register("sqlite3_with_regexp",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.RegisterFunc("REGEXP", regexp.MatchString, true)
			},
		})
}
