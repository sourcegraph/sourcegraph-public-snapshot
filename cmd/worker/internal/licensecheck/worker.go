package licensecheck

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/observation"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
)

type licenseWorker struct{}

// NewJob is the set of background jobs used for licensing enforcement and gating.
// Note: This job should only run once for a given Sourcegraph instance.
func NewJob() job.Job {
	return &licenseWorker{}
}

func (s *licenseWorker) Description() string {
	return "License check job"
}

func (*licenseWorker) Config() []env.Config {
	return nil
}

func (s *licenseWorker) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		&licenseChecksWrapper{
			logger: observationCtx.Logger,
			db:     db,
		},
	}, nil
}

type licenseChecksWrapper struct {
	logger log.Logger
	db     database.DB
}

func (l *licenseChecksWrapper) Start() {
	goroutine.Go(func() {
		licensing.StartMaxUserCount(l.logger, &usersStore{
			db: l.db,
		})
	})
	if !envvar.SourcegraphDotComMode() {
		StartLicenseCheck(context.Background(), l.logger, l.db)
	}
}

func (l *licenseChecksWrapper) Stop() {
	// no-op
}

type usersStore struct {
	db database.DB
}

func (u *usersStore) Count(ctx context.Context) (int, error) {
	return u.db.Users().Count(
		ctx,
		&database.UsersListOptions{
			ExcludeSourcegraphOperators: true,
		},
	)
}
