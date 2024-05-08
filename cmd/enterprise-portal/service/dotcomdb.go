package service

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/cloudsql"
)

func newDotComDBConn(ctx context.Context, config Config) (*dotcomdb.Database, error) {
	// TODO allow override
	if config.DotComDB.PGDSNOverride != nil {
		config, err := pgx.ParseConfig(*config.DotComDB.PGDSNOverride)
		if err != nil {
			return nil, errors.Wrap(err, "rendered PGDSN is invalid")
		}
		conn, err := pgx.ConnectConfig(ctx, config)
		if err != nil {
			return nil, err
		}
		return dotcomdb.NewDatabase(conn), nil
	}

	// Use IAM auth to connect to the Cloud SQL database.
	conn, err := cloudsql.Connect(ctx, config.DotComDB.ConnConfig)
	if err != nil {
		return nil, errors.Wrap(err, "contract.GetPostgreSQLDB")
	}
	return dotcomdb.NewDatabase(conn), nil
}
