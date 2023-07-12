package licensing

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
)

type licenseWorker struct{}

func NewLicenseWorker() job.Job {
	return &licenseWorker{}
}

func (s *licenseWorker) Description() string {
	return "Licensing jobs"
}

func (*licenseWorker) Config() []env.Config {
	return nil
}

func (s *licenseWorker) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	gs, err := db.GlobalState().Get(context.Background())
	if err != nil {
		return nil, err
	}
	return []goroutine.BackgroundRoutine{
		&licenseChecksWrapper{
			logger: observationCtx.Logger,
			siteID: gs.SiteID,
			db:     db,
		},
	}, nil
}

type licenseChecksWrapper struct {
	logger log.Logger
	siteID string
	db     database.DB
}

func (l *licenseChecksWrapper) Start() {
	goroutine.Go(func() {
		StartMaxUserCount(l.logger, &usersStore{
			db: l.db,
		})
	})
	if !envvar.SourcegraphDotComMode() {
		StartLicenseCheck(context.Background(), l.logger, l.siteID)
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
