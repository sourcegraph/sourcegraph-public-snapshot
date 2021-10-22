package discovery

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"

	"github.com/sourcegraph/sourcegraph/internal/insights"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
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
	insightStore := store.NewInsightStore(m.insights)
	loader := insights.NewLoader(m.base)
	dashboardStore := store.NewDashboardStore(m.insights)

	discovered, err := discoverIntegrated(ctx, loader)
	if err != nil {
		return err
	}

	var count, skipped, errorCount int
	for _, d := range discovered {
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

		err = migrateSeries(ctx, insightStore, d)
		if err != nil {
			// we can't do anything about errors, so we will just skip it and log it
			errorCount++
			log15.Error("insights migration: error while migrating insight", "error", err)
		}
		count++
	}
	log15.Info("insights settings migration complete", "count", count, "skipped", skipped, "errors", errorCount)

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
	return nil
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

// migrateSeries will attempt to take an insight defined in Sourcegraph settings and migrate it to the database.
func migrateSeries(ctx context.Context, insightStore *store.InsightStore, from insights.SearchInsight) (err error) {
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
			SeriesID:           Encode(timeSeries),
			Query:              timeSeries.Query,
			NextRecordingAfter: insights.NextRecording(time.Now()),
			NextSnapshotAfter:  insights.NextSnapshot(time.Now()),
		}
		var series types.InsightSeries
		// first check if this data series already exists (somebody already created an insight of this query), in which case we just need to attach the view to this data series
		existing, err := tx.GetDataSeries(ctx, store.GetDataSeriesArgs{SeriesID: Encode(timeSeries)})
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
		Title:       from.Title,
		Description: from.Description,
		UniqueID:    from.ID,
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
