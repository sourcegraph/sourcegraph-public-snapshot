package singleprogram

import (
	"bufio"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type StopPostgresFunc func() error

var noopStop = func() error { return nil }

var useEmbeddedPostgreSQL = env.MustGetBool("USE_EMBEDDED_POSTGRESQL", true, "use an embedded PostgreSQL server (to use an existing PostgreSQL server and database, set the PG* env vars)")

type postgresqlEnvVars struct {
	PGPORT, PGHOST, PGUSER, PGPASSWORD, PGDATABASE, PGSSLMODE, PGDATASOURCE string
}

func initPostgreSQL(logger log.Logger, embeddedPostgreSQLRootDir string) (StopPostgresFunc, error) {
	var vars *postgresqlEnvVars
	var stop StopPostgresFunc
	if useEmbeddedPostgreSQL {
		var err error
		stop, vars, err = startEmbeddedPostgreSQL(logger, embeddedPostgreSQLRootDir)
		if err != nil {
			return stop, errors.Wrap(err, "Failed to download or start embedded postgresql. Please start your own postgres instance then configure the PG* environment variables to connect to it as well as setting USE_EMBEDDED_POSTGRESQL=0")
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

	useSinglePostgreSQLDatabase(logger, vars)

	// Migration on startup is ideal for the app deployment because there are no other
	// simultaneously running services at startup that might interfere with a migration.
	//
	// TODO(sqs): TODO(single-binary): make this behavior more official and not just for "dev"
	setDefaultEnv(logger, "SG_DEV_MIGRATE_ON_APPLICATION_STARTUP", "1")

	return stop, nil
}

func useSinglePostgreSQLDatabase(logger log.Logger, vars *postgresqlEnvVars) {
	// Use a single PostgreSQL DB.
	//
	// For code intel:
	logger.Debug("setting CODEINTEL database variables")
	os.Setenv("CODEINTEL_PGPORT", vars.PGPORT)
	os.Setenv("CODEINTEL_PGHOST", vars.PGHOST)
	os.Setenv("CODEINTEL_PGUSER", vars.PGUSER)
	os.Setenv("CODEINTEL_PGPASSWORD", vars.PGPASSWORD)
	os.Setenv("CODEINTEL_PGDATABASE", vars.PGDATABASE)
	os.Setenv("CODEINTEL_PGSSLMODE", vars.PGSSLMODE)
	os.Setenv("CODEINTEL_PGDATASOURCE", vars.PGDATASOURCE)
	os.Setenv("CODEINTEL_PG_ALLOW_SINGLE_DB", "true")
	// And for code insights.
	logger.Debug("setting CODEINSIGHTS database variables")
	os.Setenv("CODEINSIGHTS_PGPORT", vars.PGPORT)
	os.Setenv("CODEINSIGHTS_PGHOST", vars.PGHOST)
	os.Setenv("CODEINSIGHTS_PGUSER", vars.PGUSER)
	os.Setenv("CODEINSIGHTS_PGPASSWORD", vars.PGPASSWORD)
	os.Setenv("CODEINSIGHTS_PGDATABASE", vars.PGDATABASE)
	os.Setenv("CODEINSIGHTS_PGSSLMODE", vars.PGSSLMODE)
	os.Setenv("CODEINSIGHTS_PGDATASOURCE", vars.PGDATASOURCE)
}

// debugLogLinesWriter returns an io.Writer which will log each line written to it to logger.
//
// Note: this leaks a goroutine since embedded-postgres does not provide a way
// for us to close the writer once it is finished running. In practice we only
// call this function once and postgres has the same lifetime as the process,
// so this is fine.
func debugLogLinesWriter(logger log.Logger, msg string) io.Writer {
	r, w := io.Pipe()

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			logger.Debug(msg, log.String("line", scanner.Text()))
		}
		if err := scanner.Err(); err != nil {
			logger.Error("error reading for "+msg, log.Error(err))
		}
	}()

	return w
}
