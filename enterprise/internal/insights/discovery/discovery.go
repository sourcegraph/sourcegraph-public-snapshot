package discovery

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/segmentio/ksuid"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// SettingStore is a subset of the API exposed by the database.Settings() store.
type SettingStore interface {
	GetLatest(context.Context, api.SettingsSubject) (*api.Settings, error)
	GetLastestSchemaSettings(context.Context, api.SettingsSubject) (*schema.Settings, error)
}

// InsightFilterArgs contains arguments that will filter out insights when discovered if matched.
type InsightFilterArgs struct {
	Ids []string
}

// Discover uses the given settings store to look for insights in the global user settings.
func Discover(ctx context.Context, settingStore SettingStore, loader insights.Loader, args InsightFilterArgs) ([]insights.SearchInsight, error) {
	discovered, err := discoverAll(ctx, settingStore, loader)
	if err != nil {
		return []insights.SearchInsight{}, err
	}
	return applyFilters(discovered, args), nil
}

// discoverIntegrated will load any insights that are integrated (meaning backend capable) from the extensions settings
func discoverIntegrated(ctx context.Context, loader insights.Loader) ([]insights.SearchInsight, error) {
	return loader.LoadAll(ctx)
}

func discoverAll(ctx context.Context, settingStore SettingStore, loader insights.Loader) ([]insights.SearchInsight, error) {
	// Get latest Global user settings.
	subject := api.SettingsSubject{Site: true}
	globalSettingsRaw, err := settingStore.GetLatest(ctx, subject)
	if err != nil {
		return nil, err
	}
	globalSettings, err := parseUserSettings(globalSettingsRaw)
	if err != nil {
		return nil, err
	}
	results := convertFromBackendInsight(globalSettings.Insights)
	integrated, err := discoverIntegrated(ctx, loader)
	if err != nil {
		return nil, err
	}

	return append(results, integrated...), nil
}

// convertFromBackendInsight is an adapter method that will transform the 'backend' insight schema to the schema that is
// used by the extensions on the frontend, and will be used in the future. As soon as the backend and frontend are fully integrated these
// 'backend' insights will be deprecated.
func convertFromBackendInsight(backendInsights []*schema.Insight) []insights.SearchInsight {
	converted := make([]insights.SearchInsight, 0)
	for _, backendInsight := range backendInsights {
		var temp insights.SearchInsight
		temp.Title = backendInsight.Title
		temp.Description = backendInsight.Description
		for _, series := range backendInsight.Series {
			temp.Series = append(temp.Series, insights.TimeSeries{
				Name:  series.Label,
				Query: series.Search,
			})
		}
		temp.ID = backendInsight.Id
		converted = append(converted, temp)
	}

	return converted
}

func parseUserSettings(settings *api.Settings) (*schema.Settings, error) {
	if settings == nil {
		// Settings have never been saved for this subject; equivalent to `{}`.
		return &schema.Settings{}, nil
	}
	var v schema.Settings
	if err := jsonc.Unmarshal(settings.Contents, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// applyFilters will apply any filters defined as arguments serially and return the intersection.
func applyFilters(total []insights.SearchInsight, args InsightFilterArgs) []insights.SearchInsight {
	filtered := total

	if len(args.Ids) > 0 {
		filtered = filterByIds(args.Ids, total)
	}

	return filtered
}

func filterByIds(ids []string, insight []insights.SearchInsight) []insights.SearchInsight {
	filtered := make([]insights.SearchInsight, 0)
	keys := make(map[string]bool)
	for _, id := range ids {
		keys[id] = true
	}

	for _, searchInsight := range insight {
		if _, ok := keys[searchInsight.ID]; ok {
			filtered = append(filtered, searchInsight)
		}
	}
	return filtered
}

type settingMigrator struct {
	base     dbutil.DB
	insights dbutil.DB
}

// NewMigrateSettingInsightsJob will migrate insights from settings into the database. This is a job that will be
// deprecated as soon as this functionality is available over an API.
func NewMigrateSettingInsightsJob(ctx context.Context, base dbutil.DB, insights dbutil.DB) goroutine.BackgroundRoutine {
	interval := time.Minute * 10
	m := settingMigrator{
		base:     base,
		insights: insights,
	}

	return goroutine.NewPeriodicGoroutine(ctx, interval,
		goroutine.NewHandlerWithErrorMessage("insight_setting_migrator", m.migrate))
}

func (m *settingMigrator) migrate(ctx context.Context) error {
	loader := insights.NewLoader(m.base)
	dashboardStore := store.NewDashboardStore(m.insights)

	discovered, err := discoverIntegrated(ctx, loader)
	if err != nil {
		return err
	}

	justInTimeInsights, err := insights.GetSearchInsights(ctx, m.base, insights.All)
	if err != nil {
		return errors.Wrap(err, "failed to fetch just-in-time insights from all settings")
	}

	langStatsInsights, err := insights.GetLangStatsInsights(ctx, m.base, insights.All)
	if err != nil {
		return errors.Wrap(err, "failed to fetch lang stats insights from all settings")
	}

	log15.Info("insights migration: migrating backend insights")
	m.migrateInsights(ctx, discovered, backend)

	log15.Info("insights migration: migrating frontend search insights")
	m.migrateInsights(ctx, justInTimeInsights, frontend)

	log15.Info("insights migration: migrating frontend lang stats insights")
	m.migrateLangStatsInsights(ctx, langStatsInsights)

	log15.Info("insights migration: migrating dashboards")
	dashboards, err := loader.LoadDashboards(ctx)
	if err != nil {
		return err
	}
	err = clearDashboards(ctx, m.insights)
	if err != nil {
		return errors.Wrap(err, "clearDashboards")
	}
	for _, dashboard := range dashboards {
		err := migrateDashboard(ctx, dashboardStore, dashboard)
		if err != nil {
			log15.Info("insights migration: error while migrating dashboard", "error", err)
			continue
		}
	}

	err = purgeOrphanFrontendSeries(ctx, m.insights)
	if err != nil {
		return errors.Wrap(err, "failed to purge orphaned frontend series")
	}

	return nil
}

type migrationBatch string

const (
	backend  migrationBatch = "backend"
	frontend migrationBatch = "frontend"
)

func (m *settingMigrator) migrateInsights(ctx context.Context, toMigrate []insights.SearchInsight, batch migrationBatch) {
	insightStore := store.NewInsightStore(m.insights)
	var count, skipped, errorCount int
	for _, d := range toMigrate {
		if d.ID == "" {
			// we need a unique ID, and if for some reason this insight doesn't have one, it can't be migrated.
			skipped++
			continue
		}
		err := insightStore.DeleteViewByUniqueID(ctx, d.ID)
		log15.Info("insights migration: deleting insight view", "unique_id", d.ID)
		if err != nil {
			// if we fail here there isn't much we can do in this migration, so continue
			skipped++
			continue
		}
		err = migrateSeries(ctx, insightStore, d, batch)
		if err != nil {
			// we can't do anything about errors, so we will just skip it and log it
			errorCount++
			log15.Error("insights migration: error while migrating insight", "error", err)
		}
		count++
	}
	log15.Info("insights settings migration batch complete", "batch", batch, "count", count, "skipped", skipped, "errors", errorCount)

}

func (m *settingMigrator) migrateLangStatsInsights(ctx context.Context, toMigrate []insights.LangStatsInsight) {
	insightStore := store.NewInsightStore(m.insights)
	tx, err := insightStore.Transact(ctx)
	if err != nil {
		log15.Info("insights migration: problem connecting to store, aborting lang stats migration")
		return
	}
	defer func() { err = tx.Store.Done(err) }()

	var count, skipped, errorCount int
	for _, d := range toMigrate {
		if d.ID == "" {
			// we need a unique ID, and if for some reason this insight doesn't have one, it can't be migrated.
			skipped++
			continue
		}
		err := insightStore.DeleteViewByUniqueID(ctx, d.ID)
		log15.Info("insights migration: deleting insight view", "unique_id", d.ID)
		if err != nil {
			// if we fail here there isn't much we can do in this migration, so continue
			skipped++
			continue
		}
		err = migrateLangStatSeries(ctx, insightStore, d)
		if err != nil {
			// we can't do anything about errors, so we will just skip it and log it
			errorCount++
			log15.Error("insights migration: error while migrating insight", "error", err)
		}
		count++
	}
	log15.Info("insights settings migration batch complete", "batch", "langStats", "count", count, "skipped", skipped, "errors", errorCount)
}

func migrateDashboard(ctx context.Context, dashboardStore *store.DBDashboardStore, from insights.SettingDashboard) (err error) {
	tx, err := dashboardStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Store.Done(err) }()

	dashboard := types.Dashboard{
		Title:      from.Title,
		InsightIDs: from.InsightIds,
	}
	log15.Info("insights migration: migrating dashboard", "settings_unique_id", from.ID)

	var grants []store.DashboardGrant
	if from.UserID != nil {
		grants = []store.DashboardGrant{store.UserDashboardGrant(int(*from.UserID))}
	} else if from.OrgID != nil {
		grants = []store.DashboardGrant{store.OrgDashboardGrant(int(*from.OrgID))}
	} else {
		grants = []store.DashboardGrant{store.GlobalDashboardGrant()}
	}
	_, err = dashboardStore.CreateDashboard(ctx, store.CreateDashboardArgs{Dashboard: dashboard, Grants: grants})
	if err != nil {
		return err
	}

	return nil
}

// clearDashboards will delete all dashboards. This should be deprecated as soon as possible, and is only useful to ensure a smooth migration from settings to database.
func clearDashboards(ctx context.Context, db dbutil.DB) error {
	_, err := db.ExecContext(ctx, deleteAllDashboardsSql)
	if err != nil {
		return err
	}
	return nil
}

const deleteAllDashboardsSql = `
-- source: enterprise/internal/insights/discovery/discovery.go:clearDashboards
delete from dashboard where save != true;
`

func purgeOrphanFrontendSeries(ctx context.Context, db dbutil.DB) error {
	_, err := db.ExecContext(ctx, purgeOrphanedFrontendSeries)
	if err != nil {
		return err
	}
	return nil
}

const purgeOrphanedFrontendSeries = `
-- source: enterprise/internal/insights/discovery/discovery.go:purgeOrphanFrontendSeries
with distinct_series_ids as (select distinct ivs.insight_series_id from insight_view_series ivs)
delete from insight_series
where id not in (select * from distinct_series_ids);
`

// migrateSeries will attempt to take an insight defined in Sourcegraph settings and migrate it to the database.
func migrateSeries(ctx context.Context, insightStore *store.InsightStore, from insights.SearchInsight, batch migrationBatch) (err error) {
	tx, err := insightStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Store.Done(err) }()

	log15.Info("insights migration: attempting to migrate insight", "unique_id", from.ID)
	dataSeries := make([]types.InsightSeries, len(from.Series))
	metadata := make([]types.InsightViewSeriesMetadata, len(from.Series))

	for i, timeSeries := range from.Series {
		temp := types.InsightSeries{
			Query: timeSeries.Query,
		}

		if batch == frontend {
			temp.Repositories = from.Repositories
			if temp.Repositories == nil {
				// this shouldn't be possible, but if for some reason we get here there is a malformed schema
				return errors.New("invalid schema for frontend insight, missing repositories")
			}
			interval := parseTimeInterval(from)
			temp.SampleIntervalUnit = string(interval.unit)
			temp.SampleIntervalValue = interval.value
			temp.SeriesID = ksuid.New().String() // this will cause some orphan records, but we can't use the query to match because of repo / time scope. We will purge orphan records at the end of this job.
			temp.JustInTime = true
			temp.GenerationMethod = types.Search
		} else if batch == backend {
			temp.SampleIntervalUnit = string(types.Month)
			temp.SampleIntervalValue = 1
			temp.NextRecordingAfter = insights.NextRecording(time.Now())
			temp.NextSnapshotAfter = insights.NextSnapshot(time.Now())
			temp.SeriesID = Encode(timeSeries)
			temp.JustInTime = false
			temp.GenerationMethod = types.Search
		} else {
			// not a real possibility
			return errors.Newf("invalid batch %v", batch)
		}

		var series types.InsightSeries
		// first check if this data series already exists (somebody already created an insight of this query), in which case we just need to attach the view to this data series
		existing, err := tx.GetDataSeries(ctx, store.GetDataSeriesArgs{SeriesID: temp.SeriesID})
		if err != nil {
			return errors.Wrapf(err, "unable to migrate insight unique_id: %s series_id: %s", from.ID, temp.SeriesID)
		} else if len(existing) > 0 {
			series = existing[0]
			log15.Info("insights migration: existing data series identified, attempting to construct and attach new view", "series_id", series.SeriesID, "unique_id", from.ID)
		} else {
			series, err = tx.CreateSeries(ctx, temp)
			if err != nil {
				return errors.Wrapf(err, "unable to migrate insight unique_id: %s series_id: %s", from.ID, temp.SeriesID)
			}
		}
		dataSeries[i] = series

		metadata[i] = types.InsightViewSeriesMetadata{
			Label:  timeSeries.Name,
			Stroke: timeSeries.Stroke,
		}
	}

	view := types.InsightView{
		Title:            from.Title,
		Description:      from.Description,
		UniqueID:         from.ID,
		PresentationType: types.Line,
	}

	if from.Filters != nil {
		view.Filters = types.InsightViewFilters{
			IncludeRepoRegex: from.Filters.IncludeRepoRegexp,
			ExcludeRepoRegex: from.Filters.ExcludeRepoRegexp,
		}
	}

	var grants []store.InsightViewGrant
	if from.UserID != nil {
		grants = []store.InsightViewGrant{store.UserGrant(int(*from.UserID))}
	} else if from.OrgID != nil {
		grants = []store.InsightViewGrant{store.OrgGrant(int(*from.OrgID))}
	} else {
		grants = []store.InsightViewGrant{store.GlobalGrant()}
	}

	view, err = tx.CreateView(ctx, view, grants)
	if err != nil {
		return errors.Wrapf(err, "unable to migrate insight unique_id: %s", from.ID)
	}

	for i, insightSeries := range dataSeries {
		err := tx.AttachSeriesToView(ctx, insightSeries, view, metadata[i])
		if err != nil {
			return errors.Wrapf(err, "unable to migrate insight unique_id: %s", from.ID)
		}
	}
	return nil
}

func migrateLangStatSeries(ctx context.Context, insightStore *store.InsightStore, from insights.LangStatsInsight) (err error) {
	tx, err := insightStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Store.Done(err) }()

	log15.Info("insights migration: attempting to migrate insight", "unique_id", from.ID)

	view := types.InsightView{
		Title:            from.Title,
		UniqueID:         from.ID,
		OtherThreshold:   &from.OtherThreshold,
		PresentationType: types.Pie,
	}
	series := types.InsightSeries{
		SeriesID:           ksuid.New().String(),
		Repositories:       []string{from.Repository},
		SampleIntervalUnit: string(types.Month),
	}
	var grants []store.InsightViewGrant
	if from.UserID != nil {
		grants = []store.InsightViewGrant{store.UserGrant(int(*from.UserID))}
	} else if from.OrgID != nil {
		grants = []store.InsightViewGrant{store.OrgGrant(int(*from.OrgID))}
	} else {
		grants = []store.InsightViewGrant{store.GlobalGrant()}
	}

	view, err = tx.CreateView(ctx, view, grants)
	if err != nil {
		return errors.Wrapf(err, "unable to migrate insight unique_id: %s", from.ID)
	}
	series, err = tx.CreateSeries(ctx, series)
	if err != nil {
		return errors.Wrapf(err, "unable to migrate insight unique_id: %s", from.ID)
	}
	err = tx.AttachSeriesToView(ctx, series, view, types.InsightViewSeriesMetadata{})
	if err != nil {
		return errors.Wrapf(err, "unable to migrate insight unique_id: %s", from.ID)
	}

	return nil
}

// there seems to be some global insights with possibly old schema that have a step field
func parseTimeInterval(insight insights.SearchInsight) timeInterval {
	if insight.Step.Days != nil {
		return timeInterval{
			unit:  types.Day,
			value: *insight.Step.Days,
		}
	} else if insight.Step.Hours != nil {
		return timeInterval{
			unit:  types.Hour,
			value: *insight.Step.Hours,
		}
	} else if insight.Step.Weeks != nil {
		return timeInterval{
			unit:  types.Week,
			value: *insight.Step.Weeks,
		}
	} else if insight.Step.Months != nil {
		return timeInterval{
			unit:  types.Month,
			value: *insight.Step.Months,
		}
	} else if insight.Step.Years != nil {
		return timeInterval{
			unit:  types.Year,
			value: *insight.Step.Years,
		}
	} else {
		return timeInterval{
			unit:  types.Month,
			value: 1,
		}
	}
}

type timeInterval struct {
	unit  types.IntervalUnit
	value int
}
