package licensecheck

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/redispool"

	subscriptionlicensechecksv1connect "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptionlicensechecks/v1/v1connect"
)

var baseUrl = env.Get("LICENSE_CHECK_API_URL", "https://enterprise-portal.sourcegraph.com", "Base URL for license check API")

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
		var checks subscriptionlicensechecksv1connect.SubscriptionLicenseChecksServiceClient
		ep, err := url.Parse(baseUrl)
		if err == nil && ep.Host == "127.0.0.1" {
			checks = subscriptionlicensechecksv1connect.NewSubscriptionLicenseChecksServiceClient(
				httpcli.InternalDoer,
				baseUrl,
			)
		} else {
			checks = subscriptionlicensechecksv1connect.NewSubscriptionLicenseChecksServiceClient(
				httpcli.ExternalDoer,
				baseUrl,
			)
		}
		routines = append(
			routines,
			newLicenseChecker(
				context.Background(),
				observationCtx.Logger,
				db,
				redispool.Store,
				conf.DefaultClient(),
				checks,
			),
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
