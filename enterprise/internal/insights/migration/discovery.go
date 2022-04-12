// Re-purposed and copied methods from discovery and other related methods.

package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/segmentio/ksuid"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/insights"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const schemaErrorPrefix = "insights oob migration schema error"

func getLangStatsInsights(settingsRow api.Settings) []insights.LangStatsInsight {
	prefix := "codeStatsInsights."
	var raw map[string]json.RawMessage
	results := make([]insights.LangStatsInsight, 0)

	raw, err := insights.FilterSettingJson(settingsRow.Contents, prefix)
	if err != nil {
		log15.Error(schemaErrorPrefix, "owner", getOwnerName(settingsRow), "error msg", "language usage insights failed to migrate due to unrecognized schema")
		return results
	}

	for id, body := range raw {
		var temp insights.LangStatsInsight
		temp.ID = makeUniqueId(id, settingsRow.Subject)
		if err := json.Unmarshal(body, &temp); err != nil {
			log15.Error(schemaErrorPrefix, "owner", getOwnerName(settingsRow), "error msg", "language usage insight failed to migrate due to unrecognized schema")
			continue
		}
		temp.UserID = settingsRow.Subject.User
		temp.OrgID = settingsRow.Subject.Org
		results = append(results, temp)
	}

	return results
}

func getFrontendInsights(settingsRow api.Settings) []insights.SearchInsight {
	prefix := "searchInsights."
	var raw map[string]json.RawMessage
	results := make([]insights.SearchInsight, 0)

	raw, err := insights.FilterSettingJson(settingsRow.Contents, prefix)
	if err != nil {
		log15.Error(schemaErrorPrefix, "owner", getOwnerName(settingsRow), "error msg", "search insights failed to migrate due to unrecognized schema")
		return results
	}

	for id, body := range raw {
		var temp insights.SearchInsight
		temp.ID = makeUniqueId(id, settingsRow.Subject)
		if err := json.Unmarshal(body, &temp); err != nil {
			log15.Error(schemaErrorPrefix, "owner", getOwnerName(settingsRow), "error msg", "search insight failed to migrate due to unrecognized schema")
			continue
		}
		temp.UserID = settingsRow.Subject.User
		temp.OrgID = settingsRow.Subject.Org

		results = append(results, temp)
	}

	return results
}

func getBackendInsights(setting api.Settings) []insights.SearchInsight {
	prefix := "insights.allrepos"

	results := make([]insights.SearchInsight, 0)
	perms := permissionAssociations{
		userID: setting.Subject.User,
		orgID:  setting.Subject.Org,
	}

	var raw map[string]json.RawMessage
	raw, err := insights.FilterSettingJson(setting.Contents, prefix)
	if err != nil {
		log15.Error(schemaErrorPrefix, "owner", getOwnerName(setting), "error msg", "search insights failed to migrate due to unrecognized schema")
		return results
	}

	for _, val := range raw {
		// iterate for each instance of the prefix key in the settings. This should never be len > 1, but it's technically a map.
		temp := unmarshalBackendInsights(val, setting)
		if len(temp) == 0 {
			continue
		}
		results = append(results, temp.Insights(perms)...)
	}

	return results
}

func getDashboards(settingsRow api.Settings) []insights.SettingDashboard {
	prefix := "insights.dashboards"

	results := make([]insights.SettingDashboard, 0)
	var raw map[string]json.RawMessage
	raw, err := insights.FilterSettingJson(settingsRow.Contents, prefix)
	if err != nil {
		log15.Error(schemaErrorPrefix, "owner", getOwnerName(settingsRow), "error msg", "dashboards failed to migrate due to unrecognized schema")
		return results
	}
	for _, val := range raw {
		// iterate for each instance of the prefix key in the settings. This should never be len > 1, but it's technically a map.
		temp := unmarshalDashboard(val, settingsRow)
		if len(temp) == 0 {
			continue
		}
		results = append(results, temp...)
	}

	return results
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

func unmarshalBackendInsights(raw json.RawMessage, setting api.Settings) IntegratedInsights {
	var dict map[string]json.RawMessage
	result := make(IntegratedInsights)

	if err := json.Unmarshal(raw, &dict); err != nil {
		log15.Error(schemaErrorPrefix, "owner", getOwnerName(setting), "error msg", "search insights failed to migrate due to unrecognized schema")
		return result
	}

	for id, body := range dict {
		var temp insights.SearchInsight
		if err := json.Unmarshal(body, &temp); err != nil {
			log15.Error(schemaErrorPrefix, "owner", getOwnerName(setting), "error msg", "search insight failed to migrate due to unrecognized schema")
			continue
		}
		result[makeUniqueId(id, setting.Subject)] = temp
	}

	return result
}

func unmarshalDashboard(raw json.RawMessage, settingsRow api.Settings) []insights.SettingDashboard {
	var dict map[string]json.RawMessage
	result := []insights.SettingDashboard{}

	if err := json.Unmarshal(raw, &dict); err != nil {
		log15.Error(schemaErrorPrefix, "owner", getOwnerName(settingsRow), "error msg", "dashboards failed to migrate due to unrecognized schema")
		return result
	}

	for id, body := range dict {
		var temp insights.SettingDashboard
		if err := json.Unmarshal(body, &temp); err != nil {
			log15.Error(schemaErrorPrefix, "owner", getOwnerName(settingsRow), "error msg", "dashboard failed to migrate due to unrecognized schema")
			continue
		}
		temp.ID = id
		temp.UserID = settingsRow.Subject.User
		temp.OrgID = settingsRow.Subject.Org

		result = append(result, temp)
	}

	return result
}

func (m *migrator) migrateInsights(ctx context.Context, toMigrate []insights.SearchInsight, batch migrationBatch) (int, error) {
	var count int
	var errs error
	for _, d := range toMigrate {
		if d.ID == "" {
			// we need a unique ID, and if for some reason this insight doesn't have one, it can't be migrated.
			// skippable error
			count++
			log15.Error(schemaErrorPrefix, "owner", getOwnerNameFromInsight(d), "error msg", "insight failed to migrate due to missing id")
			continue
		}
		insight, err := m.insightStore.Get(ctx, store.InsightQueryArgs{UniqueID: d.ID, WithoutAuthorization: true})
		if err != nil {
			errs = errors.Append(errs, err)
			continue
		}
		if len(insight) > 0 {
			// this insight has already been migrated, so count it
			count++
			continue
		}
		err = migrateSeries(ctx, m.insightStore, m.workerBaseStore, d, batch)
		if err != nil {
			errs = errors.Append(errs, err)
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
			log15.Error(schemaErrorPrefix, "owner", getOwnerNameFromLangStatsInsight(d), "error msg", "insight failed to migrate due to missing id")
			count++
			continue
		}
		insight, err := m.insightStore.Get(ctx, store.InsightQueryArgs{UniqueID: d.ID, WithoutAuthorization: true})
		if err != nil {
			errs = errors.Append(errs, err)
			continue
		}
		if len(insight) > 0 {
			// this insight has already been migrated, so count it towards the total
			count++
			continue
		}

		err = migrateLangStatSeries(ctx, m.insightStore, d)
		if err != nil {
			errs = errors.Append(errs, err)
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
		JustInTime:         true,
		GenerationMethod:   types.LanguageStats,
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

func migrateSeries(ctx context.Context, insightStore *store.InsightStore, workerStore *basestore.Store, from insights.SearchInsight, batch migrationBatch) (err error) {
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
				log15.Error(schemaErrorPrefix, "owner", getOwnerNameFromInsight(from), "error msg", "insight failed to migrate due to missing repositories")
				return nil
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
			temp.SeriesID = ksuid.New().String()
			temp.JustInTime = false
			temp.GenerationMethod = types.Search
		}

		var series types.InsightSeries

		// Backend series require special consideration to re-use series
		if batch == backend {
			matched, exists, err := tx.FindMatchingSeries(ctx, store.MatchSeriesArgs{
				Query:             temp.Query,
				StepIntervalUnit:  temp.SampleIntervalUnit,
				StepIntervalValue: temp.SampleIntervalValue,
			})
			if err != nil {
				return errors.Wrapf(err, "unable to migrate insight unique_id: %s series_id: %s", from.ID, temp.SeriesID)
			} else if exists {
				// If the series already exists, we can re-use that series
				series = matched
			} else {
				// If the series does not exist, we need to create a new one
				series, err = tx.CreateSeries(ctx, temp)
				if err != nil {
					return errors.Wrapf(err, "unable to migrate insight unique_id: %s series_id: %s", from.ID, temp.SeriesID)
				}

				// Also match/replace old series_points ids with the new series id
				oldId := discovery.Encode(timeSeries)
				countUpdated, silentErr := updateTimeSeriesReferences(tx.Handle().DB(), ctx, oldId, temp.SeriesID)
				if silentErr != nil {
					// If the find-replace fails, it's not a big deal. It will just need to be calcuated again.
					log15.Error("error updating series_id for series_points", "series_id", temp.SeriesID, "err", silentErr)
				} else if countUpdated == 0 {
					// If find-replace doesn't match any records, we still need to backfill, so just continue
				} else {
					// If the find-replace succeeded, we can do a similar find-replace on the jobs in the queue,
					// and then stamp the backfill_queued_at on the new series.
					silentErr = updateTimeSeriesJobReferences(workerStore, ctx, oldId, temp.SeriesID)
					if silentErr != nil {
						// If the find-replace fails, it's not a big deal. It will just need to be calcuated again.
						log15.Error("error updating series_id for jobs", "series_id", temp.SeriesID, "err", silentErr)
					} else {
						series, silentErr = tx.StampBackfill(ctx, series)
						if silentErr != nil {
							// If the stamp fails, skip it. It will just need to be calcuated again.
							log15.Error("error updating backfill_queued_at", "series_id", temp.SeriesID, "err", silentErr)
						}
					}
				}
			}
		} else {
			// If it's not a backend series, we just want to create it.
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
			log15.Error(schemaErrorPrefix, "owner", getOwnerNameFromDashboard(d), "error msg", "dashboard failed to migrate due to missing id")
			count++
			continue
		}
		err := m.migrateDashboard(ctx, d, mc)
		if err != nil {
			errs = errors.Append(errs, err)
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

func getOwnerName(settingsRow api.Settings) string {
	name := ""
	if settingsRow.Subject.User != nil {
		name = fmt.Sprintf("user id %d", *settingsRow.Subject.User)
	} else if settingsRow.Subject.Org != nil {
		name = fmt.Sprintf("org id %d", *settingsRow.Subject.Org)
	} else {
		name = "global"
	}
	return name
}

func getOwnerNameFromInsight(insight insights.SearchInsight) string {
	name := ""
	if insight.UserID != nil {
		name = fmt.Sprintf("user id %d", *insight.UserID)
	} else if insight.OrgID != nil {
		name = fmt.Sprintf("org id %d", *insight.OrgID)
	} else {
		name = "global"
	}
	return name
}

func getOwnerNameFromLangStatsInsight(insight insights.LangStatsInsight) string {
	name := ""
	if insight.UserID != nil {
		name = fmt.Sprintf("user id %d", *insight.UserID)
	} else if insight.OrgID != nil {
		name = fmt.Sprintf("org id %d", *insight.OrgID)
	} else {
		name = "global"
	}
	return name
}

func getOwnerNameFromDashboard(insight insights.SettingDashboard) string {
	name := ""
	if insight.UserID != nil {
		name = fmt.Sprintf("user id %d", *insight.UserID)
	} else if insight.OrgID != nil {
		name = fmt.Sprintf("org id %d", *insight.OrgID)
	} else {
		name = "global"
	}
	return name
}
