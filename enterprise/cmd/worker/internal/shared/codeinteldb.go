package shared

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
)

var initCodeIntelDatabaseMemo struct {
	conn *sql.DB
	err  error
	once sync.Once
}

func InitCodeIntelDatabase() (*sql.DB, error) {
	initCodeIntelDatabaseMemo.once.Do(func() {
		initCodeIntelDatabaseMemo.conn, initCodeIntelDatabaseMemo.err = initCodeIntelDatabase()
	})

	return initCodeIntelDatabaseMemo.conn, initCodeIntelDatabaseMemo.err
}

func initCodeIntelDatabase() (*sql.DB, error) {
	postgresDSN := conf.Get().ServiceConnections.CodeIntelPostgresDSN
	conf.Watch(func() {
		if newDSN := conf.Get().ServiceConnections.CodeIntelPostgresDSN; postgresDSN != newDSN {
			log.Fatalf("Detected codeintel database DSN change, restarting to take effect: %s", newDSN)
		}
	})

	db, err := dbconn.New(postgresDSN, "_codeintel")
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to codeintel database: %s", err)
	}

	if err := dbconn.MigrateDB(db, dbconn.CodeIntel); err != nil {
		return nil, fmt.Errorf("Failed to perform codeintel database migration: %s", err)
	}

	return db, nil
}
