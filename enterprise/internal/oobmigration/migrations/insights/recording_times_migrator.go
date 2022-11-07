package insights

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type recordingTimesMigrator struct {
	insightsStore *basestore.Store
}

func NewRecordingTimesMigrator() *recordingTimesMigrator {
	return &recordingTimesMigrator{}
}

var _ oobmigration.Migrator = &recordingTimesMigrator{}

func (m *recordingTimesMigrator) ID() int                 { return 17 }
func (m *recordingTimesMigrator) Interval() time.Duration { return time.Second * 10 }

func (m *recordingTimesMigrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.insightsStore.Query(ctx, sqlf.Sprintf(`
		SELECT
			CASE c2.count WHEN 0 THEN 1 ELSE
				cast(c1.count as float) / cast(c2.count as float)
			END
		FROM
			(SELECT count(*) as count FROM insight_series WHERE supports_augmentation IS TRUE) c1,
			(SELECT count(*) as count FROM insight_series) c2
	`)))
	return progress, err
}
