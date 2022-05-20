package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"

	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	migrationstore "github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type runFunc func(quiet bool, cmd ...string) (string, error)

const databaseNamePrefix = "schemadoc-gen-temp-"

var logger = log.New(os.Stderr, "", log.LstdFlags)

type databaseFactory func(dsn string, appName string, observationContext *observation.Context) (*sql.DB, error)

var schemas = map[string]struct {
	destinationFilename string
	factory             databaseFactory
}{
	"frontend":  {"schema", connections.MigrateNewFrontendDB},
	"codeintel": {"schema.codeintel", connections.MigrateNewCodeIntelDB},
	"insights":  {"schema.codeinsights", connections.MigrateNewCodeInsightsDB},
}

// This script generates markdown formatted output containing descriptions of
// the current dabase schema, obtained from postgres. The correct PGHOST,
// PGPORT, PGUSER etc. env variables must be set to run this script.
func main() {
	if err := mainErr(); err != nil {
		log.Fatal(err)
	}
}

func mainErr() error {
	dataSourcePrefix := "dbname=" + databaseNamePrefix

	g, _ := errgroup.WithContext(context.Background())
	for name, schema := range schemas {
		name, schema := name, schema
		g.Go(func() error {
			return generateAndWrite(name, schema.factory, dataSourcePrefix+name, nil, schema.destinationFilename)
		})
	}

	return g.Wait()
}

func generateAndWrite(name string, factory databaseFactory, dataSource string, commandPrefix []string, destinationFile string) error {
	run := runWithPrefix(commandPrefix)

	// Try to drop a database if it already exists
	_, _ = run(true, "dropdb", databaseNamePrefix+name)

	// Let's also try to clean up after ourselves
	defer func() { _, _ = run(true, "dropdb", databaseNamePrefix+name) }()

	if out, err := run(false, "createdb", databaseNamePrefix+name); err != nil {
		return errors.Wrap(err, fmt.Sprintf("run: %s", out))
	}

	db, err := factory(dataSource, "", &observation.TestContext)
	if err != nil {
		return err
	}
	defer db.Close()

	store := migrationstore.NewWithDB(db, "schema_migrations", migrationstore.NewOperations(&observation.TestContext))
	schemas, err := store.Describe(context.Background())
	if err != nil {
		return err
	}
	schema := schemas["public"]

	if err := os.WriteFile(destinationFile+".md", []byte(descriptions.NewPSQLFormatter().Format(schema)), os.ModePerm); err != nil {
		return err
	}

	if err := os.WriteFile(destinationFile+".json", []byte(descriptions.NewJSONFormatter().Format(schema)), os.ModePerm); err != nil {
		return err
	}

	return nil
}

func runWithPrefix(prefix []string) runFunc {
	return func(quiet bool, cmd ...string) (string, error) {
		cmd = append(prefix, cmd...)

		c := exec.Command(cmd[0], cmd[1:]...)
		if !quiet {
			c.Stderr = logger.Writer()
		}

		out, err := c.Output()
		return string(out), err
	}
}
