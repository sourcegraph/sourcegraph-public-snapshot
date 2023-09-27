pbckbge licensecheck

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"

	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
)

type licenseWorker struct{}

// NewJob is the set of bbckground jobs used for licensing enforcement bnd gbting.
// Note: This job should only run once for b given Sourcegrbph instbnce.
func NewJob() job.Job {
	return &licenseWorker{}
}

func (s *licenseWorker) Description() string {
	return "License check job"
}

func (*licenseWorker) Config() []env.Config {
	return nil
}

func (s *licenseWorker) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BbckgroundRoutine{
		&licenseChecksWrbpper{
			logger: observbtionCtx.Logger,
			db:     db,
		},
	}, nil
}

type licenseChecksWrbpper struct {
	logger log.Logger
	db     dbtbbbse.DB
}

func (l *licenseChecksWrbpper) Stbrt() {
	goroutine.Go(func() {
		licensing.StbrtMbxUserCount(l.logger, &usersStore{
			db: l.db,
		})
	})
	if !envvbr.SourcegrbphDotComMode() {
		StbrtLicenseCheck(context.Bbckground(), l.logger, l.db)
	}
}

func (l *licenseChecksWrbpper) Stop() {
	// no-op
}

type usersStore struct {
	db dbtbbbse.DB
}

func (u *usersStore) Count(ctx context.Context) (int, error) {
	return u.db.Users().Count(
		ctx,
		&dbtbbbse.UsersListOptions{
			ExcludeSourcegrbphOperbtors: true,
		},
	)
}
