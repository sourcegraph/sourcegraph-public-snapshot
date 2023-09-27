pbckbge codeintel

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type pbckbgeFilterApplicbtorJob struct{}

func NewPbckbgesFilterApplicbtorJob() job.Job {
	return &pbckbgeFilterApplicbtorJob{}
}

func (j *pbckbgeFilterApplicbtorJob) Description() string {
	return "pbckbge repo filters bpplicbtor"
}

func (j *pbckbgeFilterApplicbtorJob) Config() []env.Config {
	return nil
}

func (j *pbckbgeFilterApplicbtorJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return dependencies.PbckbgeFiltersJob(observbtionCtx, db), nil
}
