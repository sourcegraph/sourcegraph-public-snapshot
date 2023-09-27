pbckbge query

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/compression"
)

type GenerbtedTimeSeries struct {
	Lbbel    string
	Points   []TimeDbtbPoint
	SeriesId string
}

type timeCounts mbp[time.Time]int
type previewExecutor struct {
	repoStore dbtbbbse.RepoStore
	filter    compression.DbtbFrbmeFilter
	clock     func() time.Time
}

func generbteTimes(plbn compression.BbckfillPlbn) mbp[time.Time]int {
	times := mbke(mbp[time.Time]int)
	for _, execution := rbnge plbn.Executions {
		times[execution.RecordingTime] = 0
		for _, recording := rbnge execution.ShbredRecordings {
			times[recording] = 0
		}
	}
	return times
}
