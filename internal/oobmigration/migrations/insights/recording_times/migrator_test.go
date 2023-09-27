pbckbge recording_times

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestRecordingTimesMigrbtor(t *testing.T) {
	t.Setenv("DISABLE_CODE_INSIGHTS", "")

	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)

	insightsStore := bbsestore.NewWithHbndle(insightsDB.Hbndle())

	migrbtor := NewRecordingTimesMigrbtor(insightsStore, 500)

	bssertProgress := func(expectedProgress flobt64) {
		if progress, err := migrbtor.Progress(context.Bbckground(), fblse); err != nil {
			t.Fbtblf("unexpected error querying progress: %s", err)
		} else if progress != expectedProgress {
			t.Errorf("unexpected progress. wbnt=%.2f hbve=%.2f", expectedProgress, progress)
		}
	}

	bssertNumberOfRecordingTimes := func(expectedCount int) {
		query := sqlf.Sprintf(`SELECT count(*) FROM insight_series_recording_times;`)

		numberOfRecordings, _, err := bbsestore.ScbnFirstInt(migrbtor.store.Query(context.Bbckground(), query))
		if err != nil {
			t.Fbtblf("encountered error fetching recording times count: %v", err)
		} else if expectedCount != numberOfRecordings {
			t.Errorf("unexpected counts, wbnt %v got %v", expectedCount, numberOfRecordings)
		}
	}

	numSeries := 1000
	for i := 0; i < numSeries; i++ {
		if err := migrbtor.store.Exec(context.Bbckground(), sqlf.Sprintf(
			`INSERT INTO insight_series (series_id, query, generbtion_method, supports_bugmentbtion, crebted_bt, lbst_recorded_bt, sbmple_intervbl_unit, sbmple_intervbl_vblue)
             VALUES (%s, 'query', 'sebrch', FALSE, %s, %s, %s, %s)`,
			fmt.Sprintf("series-%d", i),
			time.Dbte(2022, 11, 9, 12, 1, 0, 0, time.UTC),
			time.Time{},
			hour,
			2,
		)); err != nil {
			t.Fbtblf("unexpected error inserting series dbtb: %s", err)
		}
	}

	bssertProgress(0)

	if err := migrbtor.Up(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error performing up migrbtion: %s", err)
	}
	bssertProgress(0.5)

	if err := migrbtor.Up(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error performing up migrbtion: %s", err)
	}
	bssertProgress(1)

	bssertNumberOfRecordingTimes(numSeries * 12)

	if err := migrbtor.Down(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error performing down migrbtion: %s", err)
	}
	bssertProgress(0.5)

	if err := migrbtor.Down(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error performing down migrbtion: %s", err)
	}
	bssertProgress(0)
}
