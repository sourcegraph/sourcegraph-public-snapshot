package recording_times

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

type seriesMetadata struct {
	id             int
	seriesID       string
	createdAt      time.Time
	lastRecordedAt time.Time
	interval       timeInterval
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

	series, err := selectSeriesMetadata(ctx, tx, m.batchSize)
	if err != nil {
		return errors.Wrap(err, "selectSeriesMetadata")
	}

	for id, metadata := range series {
		recordingTimes, err := selectExistingRecordingTimes(ctx, tx, metadata.seriesID)
		if err != nil {
			return errors.Wrap(err, "selectExistingRecordingTimes")
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

func selectSeriesMetadata(ctx context.Context, tx *basestore.Store, batchSize int) (map[int]seriesMetadata, error) {
	rows, err := tx.Query(ctx, sqlf.Sprintf(
		"SELECT id, series_id, created_at, last_recorded_at, sample_interval_unit, sample_interval_value FROM insight_series WHERE supports_augmentation IS FALSE ORDER BY id LIMIT %s FOR UPDATE SKIP LOCKED",
		batchSize,
	))
	if err != nil {
		return nil, err
	}

	series := make(map[int]seriesMetadata) // id -> metadata
	for rows.Next() {
		var id int
		var seriesID string
		var createdAt, lastRecordedAt time.Time
		var sampleIntervalUnit string
		var sampleIntervalValue int
		if err := rows.Scan(
			&id,
			&seriesID,
			&createdAt,
			&lastRecordedAt,
			&sampleIntervalUnit,
			&sampleIntervalValue,
		); err != nil {
			return nil, err
		}
		series[id] = seriesMetadata{
			id:             id,
			seriesID:       seriesID,
			createdAt:      createdAt,
			lastRecordedAt: lastRecordedAt,
			interval: timeInterval{
				unit:  intervalUnit(sampleIntervalUnit),
				value: sampleIntervalValue,
			},
		}
	}
	if err = basestore.CloseRows(rows, err); err != nil {
		return nil, err
	}
	return series, nil
}

func selectExistingRecordingTimes(ctx context.Context, tx *basestore.Store, seriesID string) ([]time.Time, error) {
	rows, err := tx.Query(ctx, sqlf.Sprintf(
		"SELECT DISTINCT time FROM series_points WHERE series_id = %s ORDER by time ASC", seriesID,
	))
	if err != nil {
		return nil, err
	}
	var recordingTimes []time.Time
	for rows.Next() {
		var record time.Time
		if err := rows.Scan(&record); err != nil {
			return nil, err
		}
		recordingTimes = append(recordingTimes, record)
	}
	if err = basestore.CloseRows(rows, err); err != nil {
		return nil, err
	}
	return recordingTimes, nil
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
