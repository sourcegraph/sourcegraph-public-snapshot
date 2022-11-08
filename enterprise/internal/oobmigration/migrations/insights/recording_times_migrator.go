package insights

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type recordingTimesMigrator struct {
	store *basestore.Store

	batchSize int
}

func NewRecordingTimesMigrator() *recordingTimesMigrator {
	return &recordingTimesMigrator{
		batchSize: 500,
	}
}

//var _ oobmigration.Migrator = &recordingTimesMigrator{}

func (m *recordingTimesMigrator) ID() int                 { return 17 }
func (m *recordingTimesMigrator) Interval() time.Duration { return time.Second * 10 }

func (m *recordingTimesMigrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(`
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

//func (m *recordingTimesMigrator) Up(ctx context.Context) (err error) {
//	tx, err := m.store.Transact(ctx)
//	if err != nil {
//		return err
//	}
//	defer func() { err = tx.Done(err) }()
//
//	rows, err := tx.Query(ctx, sqlf.Sprintf(
//		"SELECT id FROM insight_series WHERE supports_augmentation IS FALSE LIMIT %s FOR UPDATE SKIP LOCKED",
//		m.batchSize,
//	))
//	if err != nil {
//		return err
//	}
//	defer func() { err = basestore.CloseRows(rows, err) }()
//}
