package singleprogram

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

var useEmbeddedPostgreSQL = env.MustGetBool("USE_EMBEDDED_POSTGRESQL", true, "use an embedded PostgreSQL server (to use an existing PostgreSQL server and database, set the PG* env vars)")

type postgresqlEnvVars struct {
	PGPORT, PGHOST, PGUSER, PGPASSWORD, PGDATABASE, PGSSLMODE, PGDATASOURCE string
}

func initPostgreSQL(embeddedPostgreSQLRootDir string) error {
	var vars *postgresqlEnvVars
	if useEmbeddedPostgreSQL {
		var err error
		vars, err = startEmbeddedPostgreSQL(embeddedPostgreSQLRootDir)
		if err != nil {
			return err
		}
		os.Setenv("PGPORT", vars.PGPORT)
		os.Setenv("PGHOST", vars.PGHOST)
		os.Setenv("PGUSER", vars.PGUSER)
		os.Setenv("PGPASSWORD", vars.PGPASSWORD)
		os.Setenv("PGDATABASE", vars.PGDATABASE)
		os.Setenv("PGSSLMODE", vars.PGSSLMODE)
		os.Setenv("PGDATASOURCE", vars.PGDATASOURCE)
	} else {
		vars = &postgresqlEnvVars{
			PGPORT:       os.Getenv("PGPORT"),
			PGHOST:       os.Getenv("PGHOST"),
			PGUSER:       os.Getenv("PGUSER"),
			PGPASSWORD:   os.Getenv("PGPASSWORD"),
			PGDATABASE:   os.Getenv("PGDATABASE"),
			PGSSLMODE:    os.Getenv("PGSSLMODE"),
			PGDATASOURCE: os.Getenv("PGDATASOURCE"),
		}
	}

	useSinglePostgreSQLDatabase(vars)

	// Migration on startup is ideal for the single-program deployment because there are no other
	// simultaneously running services at startup that might interfere with a migration.
	//
	// TODO(sqs): make this behavior more official and not just for "dev"
	setDefaultEnv("SG_DEV_MIGRATE_ON_APPLICATION_STARTUP", "1")

	return nil
}

func startEmbeddedPostgreSQL(pgRootDir string) (*postgresqlEnvVars, error) {
	const pgPort = 35432
	vars := &postgresqlEnvVars{
		PGPORT:     strconv.Itoa(pgPort),
		PGHOST:     "localhost",
		PGUSER:     "sourcegraph",
		PGPASSWORD: "sourcegraph",
		PGDATABASE: "sourcegraph",
		PGSSLMODE:  "disable",
	}

	db := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Version(embeddedpostgres.V14).
			BinariesPath(filepath.Join(pgRootDir, "bin")).
			DataPath(filepath.Join(pgRootDir, "data")).
			RuntimePath(filepath.Join(pgRootDir, "runtime")).
			Username(vars.PGUSER).
			Password(vars.PGPASSWORD). // TODO(sqs) TODO!(sqs) TODO(security): autogen this
			Database(vars.PGDATABASE).
			Port(pgPort).
			Logger(log.Writer()),
	)
	if err := db.Start(); err != nil {
		return nil, err
	}
	go goroutine.MonitorBackgroundRoutines(context.Background(), &embeddedPostgreSQLBackgroundJob{db})
	return vars, nil
}

type embeddedPostgreSQLBackgroundJob struct {
	db *embeddedpostgres.EmbeddedPostgres // must be already started
}

func (db *embeddedPostgreSQLBackgroundJob) Start() {
	// Noop. We start it synchronously on purpose because everything else following it requires it.
}

func (db *embeddedPostgreSQLBackgroundJob) Stop() {
	// Sleep a short amount of time to give other services time to write to the database during their cleanup.
	time.Sleep(1000 * time.Millisecond)
	if err := db.db.Stop(); err != nil {
		fmt.Fprintln(os.Stderr, "error stopping embedded PostgreSQL:", err)
	}
}

func useSinglePostgreSQLDatabase(vars *postgresqlEnvVars) {
	// Use a single PostgreSQL DB.
	//
	// For code intel:
	setDefaultEnv("CODEINTEL_PGPORT", vars.PGPORT)
	setDefaultEnv("CODEINTEL_PGHOST", vars.PGHOST)
	setDefaultEnv("CODEINTEL_PGUSER", vars.PGUSER)
	setDefaultEnv("CODEINTEL_PGPASSWORD", vars.PGPASSWORD)
	setDefaultEnv("CODEINTEL_PGDATABASE", vars.PGDATABASE)
	setDefaultEnv("CODEINTEL_PGSSLMODE", vars.PGSSLMODE)
	setDefaultEnv("CODEINTEL_PGDATASOURCE", vars.PGDATASOURCE)
	setDefaultEnv("CODEINTEL_PG_ALLOW_SINGLE_DB", "true")
	// And for code insights.
	setDefaultEnv("CODEINSIGHTS_PGPORT", vars.PGPORT)
	setDefaultEnv("CODEINSIGHTS_PGHOST", vars.PGHOST)
	setDefaultEnv("CODEINSIGHTS_PGUSER", vars.PGUSER)
	setDefaultEnv("CODEINSIGHTS_PGPASSWORD", vars.PGPASSWORD)
	setDefaultEnv("CODEINSIGHTS_PGDATABASE", vars.PGDATABASE)
	setDefaultEnv("CODEINSIGHTS_PGSSLMODE", vars.PGSSLMODE)
	setDefaultEnv("CODEINSIGHTS_PGDATASOURCE", vars.PGDATASOURCE)
}
