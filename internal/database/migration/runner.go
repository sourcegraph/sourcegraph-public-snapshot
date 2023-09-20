package migration

import (
	"database/sql"
	"strings"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func NewRunnerWithSchemas(
	observationCtx *observation.Context,
	out *output.Output,
	appName string,
	schemaNames []string,
	schemas []*schemas.Schema,
) (*runner.Runner, error) {
	dsns, err := postgresdsn.DSNsBySchema(schemaNames)
	if err != nil {
		return nil, err
	}
	var verbose = env.LogLevel == "dbug"

	var dsnsStrings []string
	for schema, dsn := range dsns {
		dsnsStrings = append(dsnsStrings, schema+" => "+dsn)
	}
	if verbose {
		out.WriteLine(output.Linef(output.EmojiInfo, output.StyleGrey, " Connection DSNs used: %s", strings.Join(dsnsStrings, ", ")))
	}

	storeFactory := func(db *sql.DB, migrationsTable string) connections.Store {
		return connections.NewStoreShim(store.NewWithDB(observationCtx, db, migrationsTable))
	}
	r, err := connections.RunnerFromDSNsWithSchemas(out, observationCtx.Logger, dsns, appName, storeFactory, schemas)
	if err != nil {
		return nil, err
	}

	return r, nil
}
