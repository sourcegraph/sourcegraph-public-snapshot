package localstore

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/sysreq"
)

func init() {
	sysreq.AddCheck("PostgreSQL", func(ctx context.Context) (problem, fix string, err error) {
		if _, _, err := GlobalDBs(); err != nil {
			return "PostgreSQL is unavailable or misconfigured",
				"Configure the PG* env vars to connect to a PostgreSQL 9.2+ database, and initialize the database with `src pgsql create`. See https://sourcegraph.com/sourcegraph/sourcegraph/.docs/config/storage/ for more information.",
				err
		}
		return
	})
}
