package main

import (
	"database/sql"
	"log"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

func mustInitializeStore() store.Store {
	postgresDSN := conf.Get().ServiceConnections.PostgresDSN
	conf.Watch(func() {
		if newDSN := conf.Get().ServiceConnections.PostgresDSN; postgresDSN != newDSN {
			log.Fatalf("Detected database DSN change, restarting to take effect: %s", newDSN)
		}
	})

	if err := dbconn.SetupGlobalConnection(postgresDSN); err != nil {
		log.Fatalf("Failed to connect to frontend database: %s", err)
	}

	return store.NewWithDB(dbconn.Global)
}

func mustInitializeCodeIntelDatabase() *sql.DB {
	postgresDSN := conf.Get().ServiceConnections.CodeIntelPostgresDSN
	conf.Watch(func() {
		if newDSN := conf.Get().ServiceConnections.CodeIntelPostgresDSN; postgresDSN != newDSN {
			log.Fatalf("Detected codeintel database DSN change, restarting to take effect: %s", newDSN)
		}
	})

	db, err := dbconn.New(postgresDSN, "_codeintel")
	if err != nil {
		log.Fatalf("Failed to connect to codeintel database: %s", err)
	}

	if err := dbconn.MigrateDB(db, "codeintel"); err != nil {
		log.Fatalf("Failed to perform codeintel database migration: %s", err)
	}

	return db
}
