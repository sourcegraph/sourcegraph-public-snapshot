// Re-purposed and copied methods from discovery and other related methods.

package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/segmentio/ksuid"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/insights"
)

func getLangStatsInsights(ctx context.Context, settingsRow api.Settings) ([]insights.LangStatsInsight, error) {
	prefix := "codeStatsInsights."
	var raw map[string]json.RawMessage
	results := make([]insights.LangStatsInsight, 0)

	raw, err := insights.FilterSettingJson(settingsRow.Contents, prefix)
	if err != nil {
		return nil, err
	}

	for id, body := range raw {
		var temp insights.LangStatsInsight
		temp.ID = makeUniqueId(id, settingsRow.Subject)
		if err := json.Unmarshal(body, &temp); err != nil {
			// a deprecated schema collides with this field name, so skip any deserialization errors
			continue
		}
		temp.UserID = settingsRow.Subject.User
		temp.OrgID = settingsRow.Subject.Org
		results = append(results, temp)
	}

	return results, nil
}

func getFrontendInsights(ctx context.Context, settingsRow api.Settings) ([]insights.SearchInsight, error) {
	prefix := "searchInsights."
	var raw map[string]json.RawMessage
	results := make([]insights.SearchInsight, 0)

	raw, err := insights.FilterSettingJson(settingsRow.Contents, prefix)
	if err != nil {
		return nil, err
	}

	for id, body := range raw {
		var temp insights.SearchInsight
		temp.ID = makeUniqueId(id, settingsRow.Subject)
		if err := json.Unmarshal(body, &temp); err != nil {
			// a deprecated schema collides with this field name, so skip any deserialization errors
			continue
		}
		temp.UserID = settingsRow.Subject.User
		temp.OrgID = settingsRow.Subject.Org

		results = append(results, temp)
	}

	return results, nil
}

func getBackendInsights(ctx context.Context, setting api.Settings) ([]insights.SearchInsight, error) {
	prefix := "insights.allrepos"
	var multi error

	results := make([]insights.SearchInsight, 0)
	perms := permissionAssociations{
		userID: setting.Subject.User,
		orgID:  setting.Subject.Org,
	}

	var raw map[string]json.RawMessage
	raw, err := insights.FilterSettingJson(setting.Contents, prefix)
	if err != nil {
		multi = multierror.Append(multi, err)
	}

	for _, val := range raw {
		// iterate for each instance of the prefix key in the settings. This should never be len > 1, but it's technically a map.
		temp, err := unmarshalBackendInsights(val, setting)
		if err != nil {
			// this isn't actually a total failure case, we could have partially parsed this dictionary.
			multi = multierror.Append(multi, err)
		}
		results = append(results, temp.Insights(perms)...)
	}

	if multi != nil {
		log15.Error("insights: deserialization errors parsing integrated insights", "error", multi)
	}

	return results, nil
}

func getDashboards(ctx context.Context, settingsRow api.Settings) ([]insights.SettingDashboard, error) {
	prefix := "insights.dashboards"

	results := make([]insights.SettingDashboard, 0)
	var raw map[string]json.RawMessage
	raw, err := insights.FilterSettingJson(settingsRow.Contents, prefix)
	if err != nil {
		return nil, err
	}
	for _, val := range raw {
		// iterate for each instance of the prefix key in the settings. This should never be len > 1, but it's technically a map.
		temp, err := unmarshalDashboard(val, settingsRow)
		if err != nil {
			continue
		}
		results = append(results, temp...)
	}

	return results, nil
}

type permissionAssociations struct {
	userID *int32
	orgID  *int32
}

type IntegratedInsights map[string]insights.SearchInsight

func (i IntegratedInsights) Insights(perms permissionAssociations) []insights.SearchInsight {
	results := make([]insights.SearchInsight, 0)
	for key, insight := range i {
		insight.ID = key // the insight ID is the value of the dict key

		// each setting is owned by either a user or an organization, which needs to be mapped when this insight is synced
		// to preserve permissions semantics
		insight.UserID = perms.userID
		insight.OrgID = perms.orgID

		results = append(results, insight)
	}
	return results
}

func unmarshalBackendInsights(raw json.RawMessage, setting api.Settings) (IntegratedInsights, error) {
	var dict map[string]json.RawMessage
	var multi error
	result := make(IntegratedInsights)

	if err := json.Unmarshal(raw, &dict); err != nil {
		return result, err
	}

	for id, body := range dict {
		var temp insights.SearchInsight
		if err := json.Unmarshal(body, &temp); err != nil {
			multi = multierror.Append(multi, err)
			continue
		}
		result[makeUniqueId(id, setting.Subject)] = temp
	}

	return result, multi
}

func unmarshalDashboard(raw json.RawMessage, settingsRow api.Settings) ([]insights.SettingDashboard, error) {
	var dict map[string]json.RawMessage
	var multi error
	result := []insights.SettingDashboard{}

	if err := json.Unmarshal(raw, &dict); err != nil {
		return result, err
	}

	for id, body := range dict {
		var temp insights.SettingDashboard
		if err := json.Unmarshal(body, &temp); err != nil {
			multi = multierror.Append(multi, err)
			continue
		}
		temp.ID = id
		temp.UserID = settingsRow.Subject.User
		temp.OrgID = settingsRow.Subject.Org

		result = append(result, temp)
	}

	return result, multi
}

func (m *migrator) migrateInsights(ctx context.Context, toMigrate []insights.SearchInsight, batch migrationBatch) (int, error) {
	var count int
	var errs error
	for _, d := range toMigrate {
		if d.ID == "" {
			// we need a unique ID, and if for some reason this insight doesn't have one, it can't be migrated.
			// skippable error
			count++
			continue
		}
		insight, err := m.insightStore.Get(ctx, store.InsightQueryArgs{UniqueID: d.ID, WithoutAuthorization: true})
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		if len(insight) > 0 {
			// this insight has already been migrated, so count it
			count++
			continue
		}
		err = migrateSeries(ctx, m.insightStore, d, batch)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		} else {
			count++
		}
	}
	return count, errs
}

func (m *migrator) migrateLangStatsInsights(ctx context.Context, toMigrate []insights.LangStatsInsight) (int, error) {
	var count int
	var errs error
	for _, d := range toMigrate {
		if d.ID == "" {
			// we need a unique ID, and if for some reason this insight doesn't have one, it can't be migrated.
			// since it can never be migrated, we count it towards the total
			count++
			continue
		}
		insight, err := m.insightStore.Get(ctx, store.InsightQueryArgs{UniqueID: d.ID, WithoutAuthorization: true})
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		if len(insight) > 0 {
			// this insight has already been migrated, so count it towards the total
			count++
			continue
		}

		err = migrateLangStatSeries(ctx, m.insightStore, d)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		} else {
			count++
		}
	}
	return count, errs
}

func migrateLangStatSeries(ctx context.Context, insightStore *store.InsightStore, from insights.LangStatsInsight) (err error) {
	tx, err := insightStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Store.Done(err) }()

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
		return errors.Wrapf(err, "unable to migrate insight view, unique_id: %s", from.ID)
	}
	series, err = tx.CreateSeries(ctx, series)
	if err != nil {
		return errors.Wrapf(err, "unable to migrate insight series, unique_id: %s", from.ID)
	}
	err = tx.AttachSeriesToView(ctx, series, view, types.InsightViewSeriesMetadata{})
	if err != nil {
		return errors.Wrapf(err, "unable to attach series, unique_id: %s", from.ID)
	}

	return nil
}

func migrateSeries(ctx context.Context, insightStore *store.InsightStore, from insights.SearchInsight, batch migrationBatch) (err error) {
	tx, err := insightStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Store.Done(err) }()

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
				// we can't do anything to fix this, so skip this insight
				return nil
			}
			interval := parseTimeInterval(from)
			temp.SampleIntervalUnit = string(interval.unit)
			temp.SampleIntervalValue = interval.value
			temp.SeriesID = ksuid.New().String() // this will cause some orphan records, but we can't use the query to match because of repo / time scope. We will purge orphan records at the end of this job.
		} else if batch == backend {
			temp.SampleIntervalUnit = string(types.Month)
			temp.SampleIntervalValue = 1
			temp.NextRecordingAfter = insights.NextRecording(time.Now())
			temp.NextSnapshotAfter = insights.NextSnapshot(time.Now())
			temp.SeriesID = ksuid.New().String()
		}

		var series types.InsightSeries
		// first check if this data series already exists (somebody already created an insight of this query), in which case we just need to attach the view to this data series
		// existing, err := tx.GetDataSeries(ctx, store.GetDataSeriesArgs{SeriesID: temp.SeriesID})
		matched, exists, err := tx.FindMatchingSeries(ctx, store.MatchSeriesArgs{
			Query:             temp.Query,
			StepIntervalUnit:  temp.SampleIntervalUnit,
			StepIntervalValue: temp.SampleIntervalValue,
		})
		if err != nil {
			return errors.Wrapf(err, "unable to migrate insight unique_id: %s series_id: %s", from.ID, temp.SeriesID)
		} else if exists && batch == backend {
			oldId := discovery.Encode(timeSeries)
			series = matched
			log15.Info("insights migration: existing data series identified, attempting to preserve time series", "series_id", series.SeriesID, "unique_id", from.ID)
			silentErr := updateSeriesId(tx, ctx, oldId, temp.SeriesID)
			if silentErr != nil {
				// it failed - not a big deal. This will get solved if / when this series is ever updated, it will just require a recalculation.
				log15.Error("error updating series_id", "series_id", temp.SeriesID, "err", silentErr)
			} else {
				silentErr = updateTimeSeriesReferences(tx.Handle().DB(), ctx, oldId, temp.SeriesID)
				if silentErr != nil {
					// we will skip this, we can always recalculate the time series. It's okay if we had a partial failure with the
					// definition updated at the time seres not, worst case scenario we just recalculate.
					log15.Error("error migrating time series", "series_id", temp.SeriesID, "err", silentErr)
				}
			}
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

func (m *migrator) migrateDashboards(ctx context.Context, toMigrate []insights.SettingDashboard, mc migrationContext) (int, error) {
	var count int
	var errs error
	for _, d := range toMigrate {
		if d.ID == "" {
			// we need a unique ID, and if for some reason this insight doesn't have one, it can't be migrated.
			// since it can never be migrated, we count it towards the total
			count++
			continue
		}
		err := m.migrateDashboard(ctx, d, mc)
		if err != nil {
			errs = multierror.Append(errs, err)
		} else {
			count++
		}
	}
	return count, errs
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

func makeUniqueId(id string, subject api.SettingsSubject) string {
	if subject.User != nil {
		return fmt.Sprintf("%s-user-%d", id, *subject.User)
	} else if subject.Org != nil {
		return fmt.Sprintf("%s-org-%d", id, *subject.Org)
	} else {
		return id
	}
}
