package pgsql

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/pkg/sysreq"
)

func init() {
	sysreq.AddCheck("PostgreSQL", func(ctx context.Context) (problem, fix string, err error) {
		if _, err := globalDB(); err != nil {
			return "PostgreSQL is unavailable or misconfigured",
				"Configure the PG* env vars to connect to a PostgreSQL 9.2+ database, and initialize the database with `src pgsql create`.",
				err
		}
		return
	})
}
