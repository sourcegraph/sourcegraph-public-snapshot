package oobmigration

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type RegisterMigratorsFunc func(ctx context.Context, db database.DB, runner *Runner) error
