package service

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/cloudsql"
)

func newDotComDBConn(ctx context.Context, config Config) (*dotcomdb.Reader, error) {
	readerOpts := dotcomdb.ReaderOptions{
		DevOnly: !config.DotComDB.IncludeProductionLicenses,
	}

	if override := config.DotComDB.PGDSNOverride; override != nil {
		config, err := pgx.ParseConfig(*override)
		if err != nil {
			return nil, errors.Wrapf(err, "pgx.ParseConfig %q", *override)
		}
		conn, err := pgx.ConnectConfig(ctx, config)
		if err != nil {
			return nil, errors.Wrapf(err, "pgx.ConnectConfig %q", *override)
		}
		return dotcomdb.NewReader(conn, readerOpts), nil
	}

	// Use IAM auth to connect to the Cloud SQL database.
	conn, err := cloudsql.Connect(ctx, config.DotComDB.ConnConfig)
	if err != nil {
		return nil, errors.Wrap(err, "contract.GetPostgreSQLDB")
	}
	return dotcomdb.NewReader(conn, readerOpts), nil
}
