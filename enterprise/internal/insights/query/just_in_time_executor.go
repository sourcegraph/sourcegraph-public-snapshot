package query

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type GeneratedTimeSeries struct {
	Label    string
	Points   []TimeDataPoint
	SeriesId string
}

type timeCounts map[time.Time]int
type justInTimeExecutor struct {
	db        database.DB
	repoStore database.RepoStore
	filter    compression.DataFrameFilter
	clock     func() time.Time
}

func generateTimes(plan compression.BackfillPlan) map[time.Time]int {
	times := make(map[time.Time]int)
	for _, execution := range plan.Executions {
		times[execution.RecordingTime] = 0
		for _, recording := range execution.SharedRecordings {
			times[recording] = 0
		}
	}
	return times
}
