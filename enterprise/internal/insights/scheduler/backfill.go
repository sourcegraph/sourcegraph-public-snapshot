package scheduler

import "github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

type BackfillScheduler interface {
	Backfill(series types.InsightSeries) error
}
