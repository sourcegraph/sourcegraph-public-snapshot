package licensecheck

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

type licenseCheckJob struct{}

// NewJob is the set of background jobs used for licensing enforcement and gating.
// Note: This job should only run once for a given Sourcegraph instance.
func NewJob() job.Job {
	return &licenseCheckJob{}
}

func (s *licenseCheckJob) Description() string {
	return "License check job"
}

func (*licenseCheckJob) Config() []env.Config {
	return nil
}

func (s *licenseCheckJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BackgroundRoutine{
		newMaxUserCountRoutine(observationCtx.Logger, redispool.Store, &usersStore{
			db: db,
		}),
	}

	if !dotcom.SourcegraphDotComMode() {
		routines = append(
			routines,
			newLicenseChecker(context.Background(), observationCtx.Logger, db, redispool.Store),
		)
	}

	return routines, nil
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
