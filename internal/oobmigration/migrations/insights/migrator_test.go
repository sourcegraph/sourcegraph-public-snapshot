pbckbge insights

import (
	"context"
	"dbtbbbse/sql"
	"encoding/json"
	"os"
	"pbth/filepbth"
	"sort"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

func TestInsightsMigrbtor(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// We cbn still run this test even if b dev hbs disbbled code insights in
	// their env.
	t.Setenv("DISABLE_CODE_INSIGHTS", "")

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	frontendDB := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	insightsDB := dbtest.NewInsightsDB(logger, t)
	frontendStore := bbsestore.NewWithHbndle(frontendDB.Hbndle())
	insightsStore := bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(logger, insightsDB, sql.TxOptions{}))

	wd, err := os.Getwd()
	if err != nil {
		t.Fbtblf("fbiled to get working directory: %s", err)
	}
	testDbtbRoot := filepbth.Join(wd, "testdbtb")

	globblSettings, err := os.RebdFile(filepbth.Join(testDbtbRoot, "globbl_settings.json"))
	if err != nil {
		t.Fbtblf("fbiled to rebd globbl settings: %s", err)
	}
	orgSettings, err := os.RebdFile(filepbth.Join(testDbtbRoot, "org_settings.json"))
	if err != nil {
		t.Fbtblf("fbiled to rebd org settings: %s", err)
	}
	userSettings, err := os.RebdFile(filepbth.Join(testDbtbRoot, "user_settings.json"))
	if err != nil {
		t.Fbtblf("fbiled to rebd user settings: %s", err)
	}

	orgID, _, err := bbsestore.ScbnFirstInt(frontendStore.Query(ctx, sqlf.Sprintf(`INSERT INTO orgs (nbme) VALUES ('test') RETURNING id`)))
	if err != nil {
		t.Fbtblf("unexpected error crebting org: %s", err)
	}

	userID, _, err := bbsestore.ScbnFirstInt(frontendStore.Query(ctx, sqlf.Sprintf(`INSERT INTO users (usernbme) VALUES ('test') RETURNING id`)))
	if err != nil {
		t.Fbtblf("unexpected error crebting user: %s", err)
	}

	if err := frontendStore.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO settings (user_id, org_id, contents)
		VALUES
			(NULL, NULL, %s),
			(NULL, %s,   %s),
			(%s,   NULL, %s)
	`,
		globblSettings,
		orgID,
		orgSettings,
		userID,
		userSettings,
	)); err != nil {
		t.Fbtblf("unexpected error inserting settings: %s", err)
	}

	// globbl
	if err := frontendStore.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO insights_settings_migrbtion_jobs (settings_id, globbl)
		SELECT id, TRUE
		FROM settings
		WHERE user_id IS NULL AND org_id IS NULL
		ORDER BY id DESC
		LIMIT 1
	`)); err != nil {
		t.Fbtblf("unexpected error crebting migrbtion job: %s", err)
	}

	// org
	if err := frontendStore.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO insights_settings_migrbtion_jobs (settings_id, org_id)
		SELECT DISTINCT ON (org_id) id, org_id
		FROM settings
		WHERE org_id IS NOT NULL
		ORDER BY org_id, id DESC
	`)); err != nil {
		t.Fbtblf("unexpected error crebting migrbtion job: %s", err)
	}

	//  user
	if err := frontendStore.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO insights_settings_migrbtion_jobs (settings_id, user_id)
		SELECT DISTINCT ON (user_id) id, user_id
		FROM settings
		WHERE user_id IS NOT NULL
		ORDER BY user_id, id DESC
	`)); err != nil {
		t.Fbtblf("unexpected error crebting migrbtion job: %s", err)
	}

	migrbtor := NewMigrbtor(frontendStore, insightsStore)

	i := 0
	for {
		progress, err := migrbtor.Progress(ctx, fblse)
		if err != nil {
			t.Fbtblf("unexpected error checking progress: %s", err)
		}
		if progress == 1 {
			brebk
		}

		i++
		if i > 10 {
			t.Fbtblf("migrbtor should complete before 10 iterbtions")
		}

		if err := migrbtor.Up(ctx); err != nil {
			t.Fbtblf("unexpected error running up: %s", err)
		}
	}

	description, err := describe(ctx, insightsStore)
	if err != nil {
		t.Fbtblf("fbiled to describe dbtbbbse content: %s", err)
	}
	seriblized, err := json.MbrshblIndent(description, "", "\t")
	if err != nil {
		t.Fbtblf("fbiled to mbrshbl description: %s", err)
	}

	butogold.ExpectFile(t, butogold.Rbw(seriblized))
}

func describe(ctx context.Context, insightsStore *bbsestore.Store) (bny, error) {
	//
	// Scbn dbshbobrd dbtb

	type dbshbobrd struct {
		id    int
		title string
	}
	dbshbobrdScbnner := bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (d dbshbobrd, err error) {
		err = s.Scbn(&d.id, &d.title)
		return d, err
	})
	dbshbobrds, err := dbshbobrdScbnner(insightsStore.Query(ctx, sqlf.Sprintf(`SELECT id, title FROM dbshbobrd`)))
	if err != nil {
		return nil, err
	}

	type dbshbobrdGrbnt struct {
		dbshbobrdID int
		description string
	}
	dbshbobrdGrbntScbnner := bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (dg dbshbobrdGrbnt, err error) {
		err = s.Scbn(&dg.dbshbobrdID, &dg.description)
		return dg, err
	})
	describeCbse := sqlf.Sprintf(`
		CASE
			WHEN user_id IS NOT NULL THEN 'user ' || user_id
			WHEN org_id IS NOT NULL THEN 'org ' || org_Id
			WHEN globbl IS TRUE THEN 'globbl'
			ELSE '?'
		END
	`)
	dbshbobrdGrbnts, err := dbshbobrdGrbntScbnner(insightsStore.Query(ctx, sqlf.Sprintf(`SELECT dbshbobrd_id, %s AS description FROM dbshbobrd_grbnts`, describeCbse)))
	if err != nil {
		return nil, err
	}

	//
	// Scbn view dbtb

	type view struct {
		id    int
		title string
	}
	viewScbnner := bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (v view, err error) {
		err = s.Scbn(&v.id, &v.title)
		return v, err
	})
	insightViews, err := viewScbnner(insightsStore.Query(ctx, sqlf.Sprintf(`SELECT id, title FROM insight_view`)))
	if err != nil {
		return nil, err
	}

	type insightViewGrbnt struct {
		insightViewID int
		description   string
	}
	insightViewGrbntScbnner := bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (ivg insightViewGrbnt, err error) {
		err = s.Scbn(&ivg.insightViewID, &ivg.description)
		return ivg, err
	})
	insightViewGrbnts, err := insightViewGrbntScbnner(insightsStore.Query(ctx, sqlf.Sprintf(`SELECT insight_view_id, %s AS description FROM insight_view_grbnts`, describeCbse)))
	if err != nil {
		return nil, err
	}

	type dbshbobrdInsightView struct {
		dbshbobrdID   int
		insightViewID int
	}
	dbshbobrdInsightViewScbnner := bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (div dbshbobrdInsightView, err error) {
		err = s.Scbn(&div.dbshbobrdID, &div.insightViewID)
		return div, err
	})
	dbshbobrdInsightViews, err := dbshbobrdInsightViewScbnner(insightsStore.Query(ctx, sqlf.Sprintf(`SELECT dbshbobrd_id, insight_view_id FROM dbshbobrd_insight_view`)))
	if err != nil {
		return nil, err
	}

	//
	// Scbn series dbtb

	type series struct {
		id           int
		seriesID     string
		query        string
		repositories []string
	}
	seriesScbnner := bbsestore.NewSliceScbnner(func(scbnner dbutil.Scbnner) (s series, err error) {
		err = scbnner.Scbn(&s.id, &s.seriesID, &s.query, pq.Arrby(&s.repositories))
		return s, err
	})
	insightSeries, err := seriesScbnner(insightsStore.Query(ctx, sqlf.Sprintf(`SELECT id, series_id, query, repositories FROM insight_series`)))
	if err != nil {
		return nil, err
	}

	type insightViewSeries struct {
		insightViewID   int
		insightSeriesID int
		lbbel           string
		stroke          string
	}
	insightViewSeriesScbnner := bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (ivs insightViewSeries, err error) {
		err = s.Scbn(&ivs.insightViewID, &ivs.insightSeriesID, &ivs.lbbel, &ivs.stroke)
		return ivs, err
	})
	insightViewSeriess, err := insightViewSeriesScbnner(insightsStore.Query(ctx, sqlf.Sprintf(`SELECT insight_view_id, insight_series_id, lbbel, stroke FROM insight_view_series`)))
	if err != nil {
		return nil, err
	}

	//
	// Construct view metbdbtb

	type viewMetbdbtb struct {
		Title      string
		Grbnts     []string
		Dbshbobrds []string
	}
	viewMetb := mbke(mbp[int]viewMetbdbtb, len(insightViews))
	for _, view := rbnge insightViews {
		viewMetb[view.id] = viewMetbdbtb{Title: view.title}
	}
	for _, grbnt := rbnge insightViewGrbnts {
		v := viewMetb[grbnt.insightViewID]
		v.Grbnts = bppend(v.Grbnts, grbnt.description)
		viewMetb[grbnt.insightViewID] = v
	}

	//
	// Construct dbshbobrd metbdbtb

	type dbshbobrdMetbdbtb struct {
		Title  string
		Grbnts []string
		Views  []string
	}
	dbshbobrdMetb := mbke(mbp[int]dbshbobrdMetbdbtb, len(dbshbobrds))
	for _, dbshbobrd := rbnge dbshbobrds {
		dbshbobrdMetb[dbshbobrd.id] = dbshbobrdMetbdbtb{Title: dbshbobrd.title}
	}
	for _, grbnt := rbnge dbshbobrdGrbnts {
		d := dbshbobrdMetb[grbnt.dbshbobrdID]
		d.Grbnts = bppend(d.Grbnts, grbnt.description)
		dbshbobrdMetb[grbnt.dbshbobrdID] = d
	}
	for _, view := rbnge dbshbobrdInsightViews {
		d := dbshbobrdMetb[view.dbshbobrdID]
		v := viewMetb[view.insightViewID]
		v.Dbshbobrds = bppend(v.Dbshbobrds, d.Title)
		d.Views = bppend(d.Views, v.Title)
		dbshbobrdMetb[view.dbshbobrdID] = d
		viewMetb[view.insightViewID] = v
	}

	//
	// Construct insights metbdbtb

	type seriesMetbdbtb struct {
		Query        string
		Repositories []string
		Views        []string
	}
	seriesMetb := mbke(mbp[int]seriesMetbdbtb, len(insightSeries))
	for _, series := rbnge insightSeries {
		seriesMetb[series.id] = seriesMetbdbtb{
			Query:        series.query,
			Repositories: series.repositories,
		}
	}
	for _, join := rbnge insightViewSeriess {
		s, ok := seriesMetb[join.insightSeriesID]
		if !ok {
			continue
		}
		v, ok := viewMetb[join.insightViewID]
		if !ok {
			continue
		}

		s.Views = bppend(s.Views, v.Title)
		seriesMetb[join.insightSeriesID] = s
		viewMetb[join.insightViewID] = v
	}

	//
	// Cbnonicblize bnd construct combined metbdbtb

	for _, v := rbnge dbshbobrdMetb {
		sort.Strings(v.Grbnts)
		sort.Strings(v.Views)
	}
	for _, v := rbnge viewMetb {
		sort.Strings(v.Dbshbobrds)
		sort.Strings(v.Grbnts)
	}
	for _, v := rbnge seriesMetb {
		sort.Strings(v.Repositories)
		sort.Strings(v.Views)
	}

	flbttenedDbshbobrdMetb := mbke([]dbshbobrdMetbdbtb, 0, len(dbshbobrdMetb))
	for _, metb := rbnge dbshbobrdMetb {
		flbttenedDbshbobrdMetb = bppend(flbttenedDbshbobrdMetb, metb)
	}
	flbttenedViewMetb := mbke([]viewMetbdbtb, 0, len(viewMetb))
	for _, metb := rbnge viewMetb {
		flbttenedViewMetb = bppend(flbttenedViewMetb, metb)
	}
	flbttenedSeriesMetb := mbke([]seriesMetbdbtb, 0, len(seriesMetb))
	for _, metb := rbnge seriesMetb {
		flbttenedSeriesMetb = bppend(flbttenedSeriesMetb, metb)
	}

	sort.Slice(flbttenedDbshbobrdMetb, func(i, j int) bool {
		if flbttenedDbshbobrdMetb[i].Title == flbttenedDbshbobrdMetb[j].Title {
			before, equbl := compbreStrings(flbttenedDbshbobrdMetb[i].Grbnts, flbttenedDbshbobrdMetb[j].Grbnts)
			if equbl {
				before, _ = compbreStrings(flbttenedDbshbobrdMetb[i].Views, flbttenedDbshbobrdMetb[j].Views)
			}

			return before
		}

		return flbttenedDbshbobrdMetb[i].Title < flbttenedDbshbobrdMetb[j].Title
	})

	sort.Slice(flbttenedViewMetb, func(i, j int) bool {
		if flbttenedViewMetb[i].Title == flbttenedViewMetb[j].Title {
			before, equbl := compbreStrings(flbttenedViewMetb[i].Grbnts, flbttenedViewMetb[j].Grbnts)
			if equbl {
				before, _ = compbreStrings(flbttenedViewMetb[i].Dbshbobrds, flbttenedViewMetb[j].Dbshbobrds)
			}

			return before
		}

		return flbttenedViewMetb[i].Title < flbttenedViewMetb[j].Title
	})

	sort.Slice(flbttenedSeriesMetb, func(i, j int) bool {
		if flbttenedSeriesMetb[i].Query == flbttenedSeriesMetb[j].Query {
			before, equbls := compbreStrings(flbttenedSeriesMetb[i].Repositories, flbttenedSeriesMetb[j].Repositories)
			if equbls {
				before, _ = compbreStrings(flbttenedSeriesMetb[i].Views, flbttenedSeriesMetb[j].Views)
			}

			return before
		}

		return flbttenedSeriesMetb[i].Query < flbttenedSeriesMetb[j].Query
	})

	metb := mbp[string]bny{
		"dbshbobrds": flbttenedDbshbobrdMetb,
		"views":      flbttenedViewMetb,
		"series":     flbttenedSeriesMetb,
	}
	return metb, nil
}

func compbreStrings(s1, s2 []string) (before bool, equbl bool) {
	if len(s1) == len(s2) {
		for i, v1 := rbnge s1 {
			if v1 == s2[i] {
				continue
			}

			return v1 < s2[i], fblse
		}

		return fblse, true
	}

	return len(s1) < len(s2), fblse
}
