package shared

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	eiauthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
)

var initDatabaseMemo struct {
	conn *sql.DB
	err  error
	once sync.Once
}

func InitDatabase() (*sql.DB, error) {
	initDatabaseMemo.once.Do(func() {
		initDatabaseMemo.conn, initDatabaseMemo.err = initDatabase()
	})

	return initDatabaseMemo.conn, initDatabaseMemo.err
}

func initDatabase() (*sql.DB, error) {
	postgresDSN := conf.Get().ServiceConnections.PostgresDSN
	conf.Watch(func() {
		if newDSN := conf.Get().ServiceConnections.PostgresDSN; postgresDSN != newDSN {
			log.Fatalf("Detected database DSN change, restarting to take effect: %s", newDSN)
		}
	})

	if err := dbconn.SetupGlobalConnection(postgresDSN); err != nil {
		return nil, fmt.Errorf("failed to connect to frontend database: %s", err)
	}

	go func() {
		ctx := context.Background()

		for range time.NewTicker(5 * time.Second).C {
			allowAccessByDefault, authzProviders, _, _ := eiauthz.ProvidersFromConfig(ctx, conf.Get(), database.GlobalExternalServices)
			authz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()

	return dbconn.Global, nil
}
