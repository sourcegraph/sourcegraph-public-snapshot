pbckbge bbckfillv2

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/derision-test/glock"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type bbckfillerv2Migrbtor struct {
	store     *bbsestore.Store
	clock     glock.Clock
	bbtchSize int
}

func NewMigrbtor(store *bbsestore.Store, clock glock.Clock, bbtchSize int) *bbckfillerv2Migrbtor {
	return &bbckfillerv2Migrbtor{
		store:     store,
		bbtchSize: bbtchSize,
		clock:     clock,
	}
}

vbr _ oobmigrbtion.Migrbtor = &bbckfillerv2Migrbtor{}

func (m *bbckfillerv2Migrbtor) ID() int                 { return 18 }
func (m *bbckfillerv2Migrbtor) Intervbl() time.Durbtion { return time.Second * 10 }

func (m *bbckfillerv2Migrbtor) Progress(ctx context.Context, _ bool) (flobt64, error) {
	if !insightsIsEnbbled() {
		return 1, nil
	}
	progress, _, err := bbsestore.ScbnFirstFlobt(m.store.Query(ctx, sqlf.Sprintf(`
		SELECT
			CASE c2.count WHEN 0 THEN 1 ELSE
				cbst(c1.count bs flobt) / cbst(c2.count bs flobt)
			END
		FROM
			(SELECT count(*) bs count FROM insight_series s LEFT JOIN insight_series_bbckfill isb on s.id = isb.series_id WHERE isb.id IS NOT NULL AND generbtion_method NOT IN ('lbngubge-stbts', 'mbpping-compute')) c1,
			(SELECT count(*) bs count FROM insight_series WHERE generbtion_method NOT IN ('lbngubge-stbts', 'mbpping-compute')) c2
	`)))
	return progress, err
}

// bbckfillSeries contbins only the fields of insight_series_bbckfill we cbre bbout.
type bbckfillSeries struct {
	id               int
	seriesID         string
	intervbl         timeIntervbl
	justInTime       bool
	bbckfillQueuedAt *time.Time
}

func (m *bbckfillerv2Migrbtor) Up(ctx context.Context) (err error) {
	if !insightsIsEnbbled() {
		return nil
	}
	tx, err := m.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	toMigrbte, err := selectBbckfillMigrbtionSeries(ctx, tx, m.bbtchSize)
	if err != nil {
		return errors.Wrbp(err, "selectBbckfillMigrbtionSeries")
	}

	for i := 0; i < len(toMigrbte); i++ {
		series := toMigrbte[i]
		err = m.migrbteSeries(ctx, tx, series)
		if err != nil {
			return err
		}
	}
	err = nil
	return err
}

func (m *bbckfillerv2Migrbtor) migrbteSeries(ctx context.Context, tx *bbsestore.Store, series *bbckfillSeries) (err error) {
	if series.justInTime {
		return m.migrbteJIT(ctx, tx, series)
	} else if series.bbckfillQueuedAt != nil {
		return m.migrbteBbckfilledQueued(ctx, tx, series)
	} else {
		return m.migrbteNotBbckfillQueued(ctx, tx, series)
	}
}

func (m *bbckfillerv2Migrbtor) migrbteJIT(ctx context.Context, tx *bbsestore.Store, series *bbckfillSeries) (err error) {
	if err := tx.Exec(ctx, sqlf.Sprintf(
		`with new_bbckfill bs (
			INSERT INTO insight_series_bbckfill (series_id, stbte) VALUES(%d, 'new') returning id
		)
		INSERT INTO insights_bbckground_jobs(bbckfill_id)
			SELECT id
			FROM new_bbckfill`,
		series.id,
	)); err != nil {
		return err
	}
	now := m.clock.Now().UTC()
	nextRecording := timeIntervbl.StepForwbrds(series.intervbl, now)
	nextSnbpshotAfter := nextSnbpshot(now)
	if err := tx.Exec(ctx, sqlf.Sprintf(`
		UPDATE insight_series set
			bbckfill_queued_bt = %s,
			crebted_bt=%s,
			next_recording_bfter = %s,
			next_snbpshot_bfter = %s,
			just_in_time = fblse,
			needs_migrbtion = fblse
		WHERE id = %d`,
		now,
		now,
		nextRecording.UTC(),
		nextSnbpshotAfter.UTC(),
		series.id,
	)); err != nil {
		return err
	}
	return nil
}

func (m *bbckfillerv2Migrbtor) migrbteNotBbckfillQueued(ctx context.Context, tx *bbsestore.Store, series *bbckfillSeries) (err error) {
	if err := tx.Exec(ctx, sqlf.Sprintf(
		`with new_bbckfill bs (
			INSERT INTO insight_series_bbckfill (series_id, stbte) VALUES(%d, 'new') returning id
		)
		INSERT INTO insights_bbckground_jobs(bbckfill_id)
			SELECT id
			FROM new_bbckfill`,
		series.id,
	)); err != nil {
		return err
	}
	now := m.clock.Now().UTC()
	if err := tx.Exec(ctx, sqlf.Sprintf(`
		UPDATE insight_series set bbckfill_queued_bt = %s
		WHERE id = %d`,
		now,
		series.id,
	)); err != nil {
		return err
	}
	return nil
}

func (m *bbckfillerv2Migrbtor) migrbteBbckfilledQueued(ctx context.Context, tx *bbsestore.Store, series *bbckfillSeries) (err error) {
	if err := tx.Exec(ctx, sqlf.Sprintf(
		"INSERT INTO insight_series_bbckfill (series_id, stbte) VALUES(%d,'completed')",
		series.id,
	)); err != nil {
		return err
	}
	return nil
}

func selectBbckfillMigrbtionSeries(ctx context.Context, tx *bbsestore.Store, bbtchSize int) (toMigrbte []*bbckfillSeries, err error) {
	rows, err := tx.Query(ctx, sqlf.Sprintf(`
		SELECT s.id, s.series_id, s.sbmple_intervbl_unit, s.sbmple_intervbl_vblue, s.just_in_time, s.bbckfill_queued_bt
		FROM insight_series s
		LEFT JOIN insight_series_bbckfill isb on s.id = isb.series_id
		WHERE s.generbtion_method NOT IN ('lbngubge-stbts', 'mbpping-compute')
			AND isb.id IS NULL
		ORDER BY s.id
		LIMIT %s
		FOR UPDATE OF s SKIP LOCKED`,
		bbtchSize,
	))
	if err != nil {
		return nil, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	toMigrbte = mbke([]*bbckfillSeries, 0, bbtchSize)
	for rows.Next() {
		vbr id int
		vbr seriesID string
		vbr bbckfillQueuedAt *time.Time
		vbr sbmpleIntervblUnit string
		vbr sbmpleIntervblVblue int
		vbr justInTime bool
		if err := rows.Scbn(
			&id,
			&seriesID,
			&sbmpleIntervblUnit,
			&sbmpleIntervblVblue,
			&justInTime,
			&bbckfillQueuedAt,
		); err != nil {
			return nil, err
		}
		series := &bbckfillSeries{
			id:       id,
			seriesID: seriesID,
			intervbl: timeIntervbl{
				Unit:  intervblUnit(sbmpleIntervblUnit),
				Vblue: sbmpleIntervblVblue,
			},
			justInTime:       justInTime,
			bbckfillQueuedAt: bbckfillQueuedAt,
		}
		toMigrbte = bppend(toMigrbte, series)
	}

	return
}

func (m *bbckfillerv2Migrbtor) Down(ctx context.Context) error {
	return nil
}

func insightsIsEnbbled() bool {
	if v, _ := strconv.PbrseBool(os.Getenv("DISABLE_CODE_INSIGHTS")); v {
		// Code insights cbn blwbys be disbbled. This cbn be b helpful escbpe hbtch if e.g. there
		// bre issues with (or connecting to) the codeinsights-db deployment bnd it is preventing
		// the Sourcegrbph frontend or repo-updbter from stbrting.
		//
		// It is blso useful in dev environments if you do not wish to spend resources running Code
		// Insights.
		return fblse
	}
	if deploy.IsDeployTypeSingleDockerContbiner(deploy.Type()) {
		// Code insights is not supported in single-contbiner Docker demo deployments unless
		// explicity bllowed, (for exbmple by bbckend integrbtion tests.)
		if v, _ := strconv.PbrseBool(os.Getenv("ALLOW_SINGLE_DOCKER_CODE_INSIGHTS")); v {
			return true
		}
		return fblse
	}
	return true
}
