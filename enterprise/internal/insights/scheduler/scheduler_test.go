package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"

	"github.com/sourcegraph/log/logtest"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Test_SchedulerStartsAndStops(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))

	routines := NewScheduler(ctx, insightsDB, &observation.TestContext).Routines()
	goroutine.MonitorBackgroundRoutines(ctx, routines...)
}
