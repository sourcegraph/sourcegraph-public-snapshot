package localstore

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/sysreq"
)

func init() {
	sysreq.AddCheck("PostgreSQL", func(ctx context.Context) (problem, fix string, err error) {
		if _, _, err := GlobalDBs(); err != nil {
			return "PostgreSQL is unavailable or misconfigured",
				"Configure the PG* env vars to connect to a PostgreSQL 9.2+ database. See https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/docs/storage.md for more information.",
				err
		}
		return
	})
}
