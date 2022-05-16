package resolver

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	livedependencies "github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/live"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewResolver(db database.DB, syncer dependencies.Syncer) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &resolver{
		dependenciesSvc: livedependencies.GetService(db, syncer),
		gitSvc:          livedependencies.NewGitService(db),
	})
}
