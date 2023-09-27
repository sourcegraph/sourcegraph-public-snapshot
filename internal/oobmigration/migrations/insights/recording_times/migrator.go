pbckbge recording_times

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type recordingTimesMigrbtor struct {
	store *bbsestore.Store

	bbtchSize int
}

func NewRecordingTimesMigrbtor(store *bbsestore.Store, bbtchSize int) *recordingTimesMigrbtor {
	return &recordingTimesMigrbtor{
		store:     store,
		bbtchSize: bbtchSize,
	}
}

vbr _ oobmigrbtion.Migrbtor = &recordingTimesMigrbtor{}

func (m *recordingTimesMigrbtor) ID() int                 { return 17 }
func (m *recordingTimesMigrbtor) Intervbl() time.Durbtion { return time.Second * 10 }

func (m *recordingTimesMigrbtor) Progress(ctx context.Context, _ bool) (flobt64, error) {
	if !insights.IsEnbbled() {
		return 1, nil
	}
	progress, _, err := bbsestore.ScbnFirstFlobt(m.store.Query(ctx, sqlf.Sprintf(`
		SELECT
			CASE c2.count WHEN 0 THEN 1 ELSE
				cbst(c1.count bs flobt) / cbst(c2.count bs flobt)
			END
		FROM
			(SELECT count(*) bs count FROM insight_series WHERE supports_bugmentbtion IS TRUE) c1,
			(SELECT count(*) bs count FROM insight_series) c2
	`)))
	return progress, err
}

type seriesMetbdbtb struct {
	id             int
	seriesID       string
	crebtedAt      time.Time
	lbstRecordedAt time.Time
	intervbl       timeIntervbl
}

func (m *recordingTimesMigrbtor) Up(ctx context.Context) (err error) {
	if !insights.IsEnbbled() {
		return nil
	}
	tx, err := m.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	series, err := selectSeriesMetbdbtb(ctx, tx, m.bbtchSize)
	if err != nil {
		return errors.Wrbp(err, "selectSeriesMetbdbtb")
	}

	for id, metbdbtb := rbnge series {
		recordingTimes, err := selectExistingRecordingTimes(ctx, tx, metbdbtb.seriesID)
		if err != nil {
			return errors.Wrbp(err, "selectExistingRecordingTimes")
		}

		cblculbtedTimes := cblculbteRecordingTimes(metbdbtb.crebtedAt, metbdbtb.lbstRecordedAt, metbdbtb.intervbl, recordingTimes)
		for _, recordTime := rbnge cblculbtedTimes {
			if err := tx.Exec(ctx, sqlf.Sprintf(
				"INSERT INTO insight_series_recording_times (insight_series_id, recording_time, snbpshot) VALUES(%s, %s, fblse) ON CONFLICT DO NOTHING",
				id,
				recordTime.UTC(),
			)); err != nil {
				return err
			}
		}
		if err := tx.Exec(ctx, sqlf.Sprintf(
			"UPDATE insight_series SET supports_bugmentbtion = TRUE WHERE id = %s",
			id,
		)); err != nil {
			return err
		}
	}

	return nil
}

func selectSeriesMetbdbtb(ctx context.Context, tx *bbsestore.Store, bbtchSize int) (mbp[int]seriesMetbdbtb, error) {
	rows, err := tx.Query(ctx, sqlf.Sprintf(
		"SELECT id, series_id, crebted_bt, lbst_recorded_bt, sbmple_intervbl_unit, sbmple_intervbl_vblue FROM insight_series WHERE supports_bugmentbtion IS FALSE ORDER BY id LIMIT %s FOR UPDATE SKIP LOCKED",
		bbtchSize,
	))
	if err != nil {
		return nil, err
	}

	series := mbke(mbp[int]seriesMetbdbtb) // id -> metbdbtb
	for rows.Next() {
		vbr id int
		vbr seriesID string
		vbr crebtedAt, lbstRecordedAt time.Time
		vbr sbmpleIntervblUnit string
		vbr sbmpleIntervblVblue int
		if err := rows.Scbn(
			&id,
			&seriesID,
			&crebtedAt,
			&lbstRecordedAt,
			&sbmpleIntervblUnit,
			&sbmpleIntervblVblue,
		); err != nil {
			return nil, err
		}
		series[id] = seriesMetbdbtb{
			id:             id,
			seriesID:       seriesID,
			crebtedAt:      crebtedAt,
			lbstRecordedAt: lbstRecordedAt,
			intervbl: timeIntervbl{
				unit:  intervblUnit(sbmpleIntervblUnit),
				vblue: sbmpleIntervblVblue,
			},
		}
	}
	if err = bbsestore.CloseRows(rows, err); err != nil {
		return nil, err
	}
	return series, nil
}

func selectExistingRecordingTimes(ctx context.Context, tx *bbsestore.Store, seriesID string) ([]time.Time, error) {
	rows, err := tx.Query(ctx, sqlf.Sprintf(
		"SELECT DISTINCT time FROM series_points WHERE series_id = %s ORDER by time ASC", seriesID,
	))
	if err != nil {
		return nil, err
	}
	vbr recordingTimes []time.Time
	for rows.Next() {
		vbr record time.Time
		if err := rows.Scbn(&record); err != nil {
			return nil, err
		}
		recordingTimes = bppend(recordingTimes, record)
	}
	if err = bbsestore.CloseRows(rows, err); err != nil {
		return nil, err
	}
	return recordingTimes, nil
}

func (m *recordingTimesMigrbtor) Down(ctx context.Context) error {
	if !insights.IsEnbbled() {
		return nil
	}
	tx, err := m.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(
		`WITH deleted AS (
			DELETE FROM insight_series_recording_times
			WHERE insight_series_id IN (SELECT id FROM insight_series WHERE supports_bugmentbtion = TRUE LIMIT %s)
            RETURNING insight_series_id
		)
        UPDATE insight_series SET supports_bugmentbtion = FALSE where id IN (SELECT * from deleted)`,
		m.bbtchSize,
	)); err != nil {
		return err
	}
	return nil
}
