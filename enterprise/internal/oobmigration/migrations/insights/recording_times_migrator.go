package insights

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type recordingTimesMigrator struct {
	store *basestore.Store

	batchSize int
}

func NewRecordingTimesMigrator(store *basestore.Store, batchSize int) *recordingTimesMigrator {
	return &recordingTimesMigrator{
		store:     store,
		batchSize: batchSize,
	}
}

var _ oobmigration.Migrator = &recordingTimesMigrator{}

func (m *recordingTimesMigrator) ID() int                 { return 17 }
func (m *recordingTimesMigrator) Interval() time.Duration { return time.Second * 10 }

func (m *recordingTimesMigrator) Progress(ctx context.Context, _ bool) (float64, error) {
	if !insights.IsEnabled() {
		return 1, nil
	}
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

func (m *recordingTimesMigrator) Up(ctx context.Context) (err error) {
	if !insights.IsEnabled() {
		return nil
	}
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	rows, err := tx.Query(ctx, sqlf.Sprintf(
		"SELECT id, created_at, last_recorded_at, sample_interval_unit, sample_interval_value FROM insight_series WHERE supports_augmentation IS FALSE ORDER BY id LIMIT %s FOR UPDATE SKIP LOCKED",
		m.batchSize,
	))
	if err != nil {
		return err
	}

	series := make(map[int]seriesMetadata) // id -> metadata
	for rows.Next() {
		var id int
		var createdAt, lastRecordedAt time.Time
		var sampleIntervalUnit string
		var sampleIntervalValue int
		if err := rows.Scan(
			&id,
			&createdAt,
			&lastRecordedAt,
			&sampleIntervalUnit,
			&sampleIntervalValue,
		); err != nil {
			return err
		}
		series[id] = seriesMetadata{
			id:             id,
			createdAt:      createdAt,
			lastRecordedAt: lastRecordedAt,
			interval: timeInterval{
				unit:  intervalUnit(sampleIntervalUnit),
				value: sampleIntervalValue,
			},
		}
	}
	if err = basestore.CloseRows(rows, err); err != nil {
		return err
	}

	for id, metadata := range series {
		recordingTimesRows, err := tx.Query(ctx, sqlf.Sprintf(
			"SELECT DISTINCT recording_time FROM insight_series_recording_times WHERE insight_series_id = %s ORDER by recording_time ASC", id,
		))
		if err != nil {
			return err
		}
		var recordingTimes []time.Time
		for rows.Next() {
			var record time.Time
			if err := recordingTimesRows.Scan(&record); err != nil {
				return err
			}
			recordingTimes = append(recordingTimes, record)
		}
		if err = basestore.CloseRows(recordingTimesRows, err); err != nil {
			return err
		}

		calculatedTimes := calculateRecordingTimes(metadata.createdAt, metadata.lastRecordedAt, metadata.interval, recordingTimes)
		for _, recordTime := range calculatedTimes {
			if err := tx.Exec(ctx, sqlf.Sprintf(
				"INSERT INTO insight_series_recording_times (insight_series_id, recording_time, snapshot) VALUES(%s, %s, false) ON CONFLICT DO NOTHING",
				id,
				recordTime.UTC(),
			)); err != nil {
				return err
			}
		}
		if err := tx.Exec(ctx, sqlf.Sprintf(
			"UPDATE insight_series SET supports_augmentation = TRUE WHERE id = %s",
			id,
		)); err != nil {
			return err
		}
	}

	return nil
}

type seriesMetadata struct {
	id             int
	createdAt      time.Time
	lastRecordedAt time.Time
	interval       timeInterval
}

func (m *recordingTimesMigrator) Down(ctx context.Context) error {
	if !insights.IsEnabled() {
		return nil
	}
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(
		`WITH deleted AS (
			DELETE FROM insight_series_recording_times 
			WHERE insight_series_id IN (SELECT id FROM insight_series WHERE supports_augmentation = TRUE LIMIT %s)
            RETURNING insight_series_id
		)
        UPDATE insight_series SET supports_augmentation = FALSE where id IN (SELECT * from deleted)`,
		m.batchSize,
	)); err != nil {
		return err
	}
	return nil
}
