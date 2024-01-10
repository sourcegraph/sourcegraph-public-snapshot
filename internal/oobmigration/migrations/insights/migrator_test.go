package insights

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func TestInsightsMigrator(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// We can still run this test even if a dev has disabled code insights in
	// their env.
	t.Setenv("DISABLE_CODE_INSIGHTS", "")

	ctx := context.Background()
	logger := logtest.Scoped(t)
	frontendDB := database.NewDB(logger, dbtest.NewDB(t))
	insightsDB := dbtest.NewInsightsDB(logger, t)
	frontendStore := basestore.NewWithHandle(frontendDB.Handle())
	insightsStore := basestore.NewWithHandle(basestore.NewHandleWithDB(logger, insightsDB, sql.TxOptions{}))

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %s", err)
	}
	testDataRoot := filepath.Join(wd, "testdata")

	globalSettings, err := os.ReadFile(filepath.Join(testDataRoot, "global_settings.json"))
	if err != nil {
		t.Fatalf("failed to read global settings: %s", err)
	}
	orgSettings, err := os.ReadFile(filepath.Join(testDataRoot, "org_settings.json"))
	if err != nil {
		t.Fatalf("failed to read org settings: %s", err)
	}
	userSettings, err := os.ReadFile(filepath.Join(testDataRoot, "user_settings.json"))
	if err != nil {
		t.Fatalf("failed to read user settings: %s", err)
	}

	orgID, _, err := basestore.ScanFirstInt(frontendStore.Query(ctx, sqlf.Sprintf(`INSERT INTO orgs (name) VALUES ('test') RETURNING id`)))
	if err != nil {
		t.Fatalf("unexpected error creating org: %s", err)
	}

	userID, _, err := basestore.ScanFirstInt(frontendStore.Query(ctx, sqlf.Sprintf(`INSERT INTO users (username) VALUES ('test') RETURNING id`)))
	if err != nil {
		t.Fatalf("unexpected error creating user: %s", err)
	}

	if err := frontendStore.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO settings (user_id, org_id, contents)
		VALUES
			(NULL, NULL, %s),
			(NULL, %s,   %s),
			(%s,   NULL, %s)
	`,
		globalSettings,
		orgID,
		orgSettings,
		userID,
		userSettings,
	)); err != nil {
		t.Fatalf("unexpected error inserting settings: %s", err)
	}

	// global
	if err := frontendStore.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO insights_settings_migration_jobs (settings_id, global)
		SELECT id, TRUE
		FROM settings
		WHERE user_id IS NULL AND org_id IS NULL
		ORDER BY id DESC
		LIMIT 1
	`)); err != nil {
		t.Fatalf("unexpected error creating migration job: %s", err)
	}

	// org
	if err := frontendStore.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO insights_settings_migration_jobs (settings_id, org_id)
		SELECT DISTINCT ON (org_id) id, org_id
		FROM settings
		WHERE org_id IS NOT NULL
		ORDER BY org_id, id DESC
	`)); err != nil {
		t.Fatalf("unexpected error creating migration job: %s", err)
	}

	//  user
	if err := frontendStore.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO insights_settings_migration_jobs (settings_id, user_id)
		SELECT DISTINCT ON (user_id) id, user_id
		FROM settings
		WHERE user_id IS NOT NULL
		ORDER BY user_id, id DESC
	`)); err != nil {
		t.Fatalf("unexpected error creating migration job: %s", err)
	}

	migrator := NewMigrator(frontendStore, insightsStore)

	i := 0
	for {
		progress, err := migrator.Progress(ctx, false)
		if err != nil {
			t.Fatalf("unexpected error checking progress: %s", err)
		}
		if progress == 1 {
			break
		}

		i++
		if i > 10 {
			t.Fatalf("migrator should complete before 10 iterations")
		}

		if err := migrator.Up(ctx); err != nil {
			t.Fatalf("unexpected error running up: %s", err)
		}
	}

	description, err := describe(ctx, insightsStore)
	if err != nil {
		t.Fatalf("failed to describe database content: %s", err)
	}
	serialized, err := json.MarshalIndent(description, "", "\t")
	if err != nil {
		t.Fatalf("failed to marshal description: %s", err)
	}

	autogold.ExpectFile(t, autogold.Raw(serialized))
}

func describe(ctx context.Context, insightsStore *basestore.Store) (any, error) {
	//
	// Scan dashboard data

	type dashboard struct {
		id    int
		title string
	}
	dashboardScanner := basestore.NewSliceScanner(func(s dbutil.Scanner) (d dashboard, err error) {
		err = s.Scan(&d.id, &d.title)
		return d, err
	})
	dashboards, err := dashboardScanner(insightsStore.Query(ctx, sqlf.Sprintf(`SELECT id, title FROM dashboard`)))
	if err != nil {
		return nil, err
	}

	type dashboardGrant struct {
		dashboardID int
		description string
	}
	dashboardGrantScanner := basestore.NewSliceScanner(func(s dbutil.Scanner) (dg dashboardGrant, err error) {
		err = s.Scan(&dg.dashboardID, &dg.description)
		return dg, err
	})
	describeCase := sqlf.Sprintf(`
		CASE
			WHEN user_id IS NOT NULL THEN 'user ' || user_id
			WHEN org_id IS NOT NULL THEN 'org ' || org_Id
			WHEN global IS TRUE THEN 'global'
			ELSE '?'
		END
	`)
	dashboardGrants, err := dashboardGrantScanner(insightsStore.Query(ctx, sqlf.Sprintf(`SELECT dashboard_id, %s AS description FROM dashboard_grants`, describeCase)))
	if err != nil {
		return nil, err
	}

	//
	// Scan view data

	type view struct {
		id    int
		title string
	}
	viewScanner := basestore.NewSliceScanner(func(s dbutil.Scanner) (v view, err error) {
		err = s.Scan(&v.id, &v.title)
		return v, err
	})
	insightViews, err := viewScanner(insightsStore.Query(ctx, sqlf.Sprintf(`SELECT id, title FROM insight_view`)))
	if err != nil {
		return nil, err
	}

	type insightViewGrant struct {
		insightViewID int
		description   string
	}
	insightViewGrantScanner := basestore.NewSliceScanner(func(s dbutil.Scanner) (ivg insightViewGrant, err error) {
		err = s.Scan(&ivg.insightViewID, &ivg.description)
		return ivg, err
	})
	insightViewGrants, err := insightViewGrantScanner(insightsStore.Query(ctx, sqlf.Sprintf(`SELECT insight_view_id, %s AS description FROM insight_view_grants`, describeCase)))
	if err != nil {
		return nil, err
	}

	type dashboardInsightView struct {
		dashboardID   int
		insightViewID int
	}
	dashboardInsightViewScanner := basestore.NewSliceScanner(func(s dbutil.Scanner) (div dashboardInsightView, err error) {
		err = s.Scan(&div.dashboardID, &div.insightViewID)
		return div, err
	})
	dashboardInsightViews, err := dashboardInsightViewScanner(insightsStore.Query(ctx, sqlf.Sprintf(`SELECT dashboard_id, insight_view_id FROM dashboard_insight_view`)))
	if err != nil {
		return nil, err
	}

	//
	// Scan series data

	type series struct {
		id           int
		seriesID     string
		query        string
		repositories []string
	}
	seriesScanner := basestore.NewSliceScanner(func(scanner dbutil.Scanner) (s series, err error) {
		err = scanner.Scan(&s.id, &s.seriesID, &s.query, pq.Array(&s.repositories))
		return s, err
	})
	insightSeries, err := seriesScanner(insightsStore.Query(ctx, sqlf.Sprintf(`SELECT id, series_id, query, repositories FROM insight_series`)))
	if err != nil {
		return nil, err
	}

	type insightViewSeries struct {
		insightViewID   int
		insightSeriesID int
		label           string
		stroke          string
	}
	insightViewSeriesScanner := basestore.NewSliceScanner(func(s dbutil.Scanner) (ivs insightViewSeries, err error) {
		err = s.Scan(&ivs.insightViewID, &ivs.insightSeriesID, &ivs.label, &ivs.stroke)
		return ivs, err
	})
	insightViewSeriess, err := insightViewSeriesScanner(insightsStore.Query(ctx, sqlf.Sprintf(`SELECT insight_view_id, insight_series_id, label, stroke FROM insight_view_series`)))
	if err != nil {
		return nil, err
	}

	//
	// Construct view metadata

	type viewMetadata struct {
		Title      string
		Grants     []string
		Dashboards []string
	}
	viewMeta := make(map[int]viewMetadata, len(insightViews))
	for _, view := range insightViews {
		viewMeta[view.id] = viewMetadata{Title: view.title}
	}
	for _, grant := range insightViewGrants {
		v := viewMeta[grant.insightViewID]
		v.Grants = append(v.Grants, grant.description)
		viewMeta[grant.insightViewID] = v
	}

	//
	// Construct dashboard metadata

	type dashboardMetadata struct {
		Title  string
		Grants []string
		Views  []string
	}
	dashboardMeta := make(map[int]dashboardMetadata, len(dashboards))
	for _, dashboard := range dashboards {
		dashboardMeta[dashboard.id] = dashboardMetadata{Title: dashboard.title}
	}
	for _, grant := range dashboardGrants {
		d := dashboardMeta[grant.dashboardID]
		d.Grants = append(d.Grants, grant.description)
		dashboardMeta[grant.dashboardID] = d
	}
	for _, view := range dashboardInsightViews {
		d := dashboardMeta[view.dashboardID]
		v := viewMeta[view.insightViewID]
		v.Dashboards = append(v.Dashboards, d.Title)
		d.Views = append(d.Views, v.Title)
		dashboardMeta[view.dashboardID] = d
		viewMeta[view.insightViewID] = v
	}

	//
	// Construct insights metadata

	type seriesMetadata struct {
		Query        string
		Repositories []string
		Views        []string
	}
	seriesMeta := make(map[int]seriesMetadata, len(insightSeries))
	for _, series := range insightSeries {
		seriesMeta[series.id] = seriesMetadata{
			Query:        series.query,
			Repositories: series.repositories,
		}
	}
	for _, join := range insightViewSeriess {
		s, ok := seriesMeta[join.insightSeriesID]
		if !ok {
			continue
		}
		v, ok := viewMeta[join.insightViewID]
		if !ok {
			continue
		}

		s.Views = append(s.Views, v.Title)
		seriesMeta[join.insightSeriesID] = s
		viewMeta[join.insightViewID] = v
	}

	//
	// Canonicalize and construct combined metadata

	for _, v := range dashboardMeta {
		sort.Strings(v.Grants)
		sort.Strings(v.Views)
	}
	for _, v := range viewMeta {
		sort.Strings(v.Dashboards)
		sort.Strings(v.Grants)
	}
	for _, v := range seriesMeta {
		sort.Strings(v.Repositories)
		sort.Strings(v.Views)
	}

	flattenedDashboardMeta := make([]dashboardMetadata, 0, len(dashboardMeta))
	for _, meta := range dashboardMeta {
		flattenedDashboardMeta = append(flattenedDashboardMeta, meta)
	}
	flattenedViewMeta := make([]viewMetadata, 0, len(viewMeta))
	for _, meta := range viewMeta {
		flattenedViewMeta = append(flattenedViewMeta, meta)
	}
	flattenedSeriesMeta := make([]seriesMetadata, 0, len(seriesMeta))
	for _, meta := range seriesMeta {
		flattenedSeriesMeta = append(flattenedSeriesMeta, meta)
	}

	sort.Slice(flattenedDashboardMeta, func(i, j int) bool {
		if flattenedDashboardMeta[i].Title == flattenedDashboardMeta[j].Title {
			before, equal := compareStrings(flattenedDashboardMeta[i].Grants, flattenedDashboardMeta[j].Grants)
			if equal {
				before, _ = compareStrings(flattenedDashboardMeta[i].Views, flattenedDashboardMeta[j].Views)
			}

			return before
		}

		return flattenedDashboardMeta[i].Title < flattenedDashboardMeta[j].Title
	})

	sort.Slice(flattenedViewMeta, func(i, j int) bool {
		if flattenedViewMeta[i].Title == flattenedViewMeta[j].Title {
			before, equal := compareStrings(flattenedViewMeta[i].Grants, flattenedViewMeta[j].Grants)
			if equal {
				before, _ = compareStrings(flattenedViewMeta[i].Dashboards, flattenedViewMeta[j].Dashboards)
			}

			return before
		}

		return flattenedViewMeta[i].Title < flattenedViewMeta[j].Title
	})

	sort.Slice(flattenedSeriesMeta, func(i, j int) bool {
		if flattenedSeriesMeta[i].Query == flattenedSeriesMeta[j].Query {
			before, equals := compareStrings(flattenedSeriesMeta[i].Repositories, flattenedSeriesMeta[j].Repositories)
			if equals {
				before, _ = compareStrings(flattenedSeriesMeta[i].Views, flattenedSeriesMeta[j].Views)
			}

			return before
		}

		return flattenedSeriesMeta[i].Query < flattenedSeriesMeta[j].Query
	})

	meta := map[string]any{
		"dashboards": flattenedDashboardMeta,
		"views":      flattenedViewMeta,
		"series":     flattenedSeriesMeta,
	}
	return meta, nil
}

func compareStrings(s1, s2 []string) (before bool, equal bool) {
	if len(s1) == len(s2) {
		for i, v1 := range s1 {
			if v1 == s2[i] {
				continue
			}

			return v1 < s2[i], false
		}

		return false, true
	}

	return len(s1) < len(s2), false
}
