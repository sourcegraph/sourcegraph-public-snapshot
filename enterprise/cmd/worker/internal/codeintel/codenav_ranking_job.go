package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/background/ranking"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type rankingGraphSerializerJob struct{}

func NewRankingGraphSerializerJob() job.Job {
	return &rankingGraphSerializerJob{}
}

func (j *rankingGraphSerializerJob) Description() string {
	return ""
}

func (j *rankingGraphSerializerJob) Config() []env.Config {
	return []env.Config{
		ranking.ConfigInst,
	}
}

func (j *rankingGraphSerializerJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices()
	if err != nil {
		return nil, err
	}

	return ranking.NewGraphSerializers(codenav.GetBackgroundJobs(services.CodenavService)), nil
}
