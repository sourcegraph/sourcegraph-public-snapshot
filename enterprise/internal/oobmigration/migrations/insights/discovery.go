// Re-purposed and copied methods from discovery and other related methods.

package insights

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/segmentio/ksuid"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const schemaErrorPrefix = "insights oob migration schema error"

// filterSettingJson will return a json map that only contains keys that match a prefix string, mapped to the keyed contents.
func filterSettingJson(settingJson string, prefix string) (map[string]json.RawMessage, error) {
	var raw map[string]json.RawMessage

	if err := jsonc.Unmarshal(settingJson, &raw); err != nil {
		return map[string]json.RawMessage{}, err
	}

	filtered := make(map[string]json.RawMessage)
	for key, val := range raw {
		if strings.HasPrefix(key, prefix) {
			filtered[key] = val
		}
	}

	return filtered, nil
}

func getLangStatsInsights(settingsRow settings) []langStatsInsight {
	prefix := "codeStatsInsights."
	var raw map[string]json.RawMessage
	results := make([]langStatsInsight, 0)

	raw, err := filterSettingJson(settingsRow.Contents, prefix)
	if err != nil {
		log15.Error(schemaErrorPrefix, "owner", getOwnerName(settingsRow), "error msg", "language usage insights failed to migrate due to unrecognized schema")
		return results
	}

	for id, body := range raw {
		var temp langStatsInsight
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

func getFrontendInsights(settingsRow settings) []searchInsight {
	prefix := "searchInsights."
	var raw map[string]json.RawMessage
	results := make([]searchInsight, 0)

	raw, err := filterSettingJson(settingsRow.Contents, prefix)
	if err != nil {
		log15.Error(schemaErrorPrefix, "owner", getOwnerName(settingsRow), "error msg", "search insights failed to migrate due to unrecognized schema")
		return results
	}

	for id, body := range raw {
		var temp searchInsight
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

func getBackendInsights(setting settings) []searchInsight {
	prefix := "insights.allrepos"

	results := make([]searchInsight, 0)
	perms := permissionAssociations{
		userID: setting.Subject.User,
		orgID:  setting.Subject.Org,
	}

	var raw map[string]json.RawMessage
	raw, err := filterSettingJson(setting.Contents, prefix)
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

func getDashboards(settingsRow settings) []settingDashboard {
	prefix := "insights.dashboards"

	results := make([]settingDashboard, 0)
	var raw map[string]json.RawMessage
	raw, err := filterSettingJson(settingsRow.Contents, prefix)
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

func (i integratedInsights) Insights(perms permissionAssociations) []searchInsight {
	results := make([]searchInsight, 0)
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

func unmarshalBackendInsights(raw json.RawMessage, setting settings) integratedInsights {
	var dict map[string]json.RawMessage
	result := make(integratedInsights)

	if err := json.Unmarshal(raw, &dict); err != nil {
		log15.Error(schemaErrorPrefix, "owner", getOwnerName(setting), "error msg", "search insights failed to migrate due to unrecognized schema")
		return result
	}

	for id, body := range dict {
		var temp searchInsight
		if err := json.Unmarshal(body, &temp); err != nil {
			log15.Error(schemaErrorPrefix, "owner", getOwnerName(setting), "error msg", "search insight failed to migrate due to unrecognized schema")
			continue
		}
		result[makeUniqueId(id, setting.Subject)] = temp
	}

	return result
}

func unmarshalDashboard(raw json.RawMessage, settingsRow settings) []settingDashboard {
	var dict map[string]json.RawMessage
	result := []settingDashboard{}

	if err := json.Unmarshal(raw, &dict); err != nil {
		log15.Error(schemaErrorPrefix, "owner", getOwnerName(settingsRow), "error msg", "dashboards failed to migrate due to unrecognized schema")
		return result
	}

	for id, body := range dict {
		var temp settingDashboard
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

func (m *migrator) migrateInsights(ctx context.Context, toMigrate []searchInsight, batch migrationBatch) (int, error) {
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

		numInsights, _, err := basestore.ScanFirstInt(m.insightsStore.Query(ctx, sqlf.Sprintf(`
			SELECT COUNT(*)
			FROM (
				SELECT *
				FROM insight_view
				WHERE unique_id = %s
				ORDER BY unique_id
			) iv
			JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
			JOIN insight_series i ON ivs.insight_series_id = i.id
			WHERE i.deleted_at IS NULL
		`,
			d.ID,
		)))
		if err != nil {
			errs = errors.Append(errs, err)
			continue
		}
		if numInsights > 0 {
			// this insight has already been migrated, so count it
			count++
			continue
		}
		err = migrateSeries(ctx, m.insightsStore, m.frontendStore, d, batch)
		if err != nil {
			errs = errors.Append(errs, err)
			continue
		} else {
			count++
		}
	}
	return count, errs
}

func (m *migrator) migrateLangStatsInsights(ctx context.Context, toMigrate []langStatsInsight) (int, error) {
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
		numInsights, _, err := basestore.ScanFirstInt(m.insightsStore.Query(ctx, sqlf.Sprintf(`
			SELECT COUNT(*)
			FROM (
				SELECT *
				FROM insight_view
				WHERE unique_id = %s
				ORDER BY unique_id
			) iv
			JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
			JOIN insight_series i ON ivs.insight_series_id = i.id
			WHERE i.deleted_at IS NULL
		`,
			d.ID,
		)))
		if err != nil {
			errs = errors.Append(errs, err)
			continue
		}
		if numInsights > 0 {
			// this insight has already been migrated, so count it towards the total
			count++
			continue
		}

		err = migrateLangStatSeries(ctx, m.insightsStore, d)
		if err != nil {
			errs = errors.Append(errs, err)
			continue
		} else {
			count++
		}
	}
	return count, errs
}

func userGrant(userID int) insightViewGrant {
	return insightViewGrant{UserID: &userID}
}

func orgGrant(orgID int) insightViewGrant {
	return insightViewGrant{OrgID: &orgID}
}

func globalGrant() insightViewGrant {
	b := true
	return insightViewGrant{Global: &b}
}

// nextRecording calculates the time that a series recording should occur given the current or most recent recording time.
func nextRecording(current time.Time) time.Time {
	year, month, _ := current.In(time.UTC).Date()
	return time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
}

func nextSnapshot(current time.Time) time.Time {
	year, month, day := current.In(time.UTC).Date()
	return time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)
}

func migrateLangStatSeries(ctx context.Context, insightStore *basestore.Store, from langStatsInsight) (err error) {
	tx, err := insightStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	now := time.Now()
	view := insightView{
		Title:            from.Title,
		UniqueID:         from.ID,
		OtherThreshold:   &from.OtherThreshold,
		PresentationType: Pie,
	}
	series := insightSeries{
		SeriesID:           ksuid.New().String(),
		Repositories:       []string{from.Repository},
		SampleIntervalUnit: string(Month),
		JustInTime:         true,
		GenerationMethod:   LanguageStats,
		CreatedAt:          now,
	}
	var grants []insightViewGrant
	if from.UserID != nil {
		grants = []insightViewGrant{userGrant(int(*from.UserID))}
	} else if from.OrgID != nil {
		grants = []insightViewGrant{orgGrant(int(*from.OrgID))}
	} else {
		grants = []insightViewGrant{globalGrant()}
	}

	viewID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(`
	INSERT INTO insight_view (
		title,
		description,
		unique_id,
		default_filter_include_repo_regex,
		default_filter_exclude_repo_regex,
		default_filter_search_contexts,
		other_threshold,
		presentation_type,
	)
	VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
	RETURNING id
	`,
		view.Title,
		view.Description,
		view.UniqueID,
		view.Filters.IncludeRepoRegex,
		view.Filters.ExcludeRepoRegex,
		pq.Array(view.Filters.SearchContexts),
		view.OtherThreshold,
		view.PresentationType,
	)))
	if err != nil {
		return errors.Wrapf(err, "unable to migrate insight view, unique_id: %s", from.ID)
	}
	view.ID = viewID
	values := make([]*sqlf.Query, 0, len(grants))
	for _, grant := range grants {
		values = append(values, sqlf.Sprintf("(%s, %s, %s, %s)", view.ID, grant.OrgID, grant.UserID, grant.Global))
	}
	err = tx.Exec(ctx, sqlf.Sprintf(`INSERT INTO insight_view_grants (insight_view_id, org_id, user_id, global) VALUES %s`, sqlf.Join(values, ", ")))
	if err != nil {
		return errors.Wrapf(err, "unable to migrate insight view, unique_id: %s", from.ID)
	}

	interval := TimeInterval{
		Unit:  intervalUnit(series.SampleIntervalUnit),
		Value: series.SampleIntervalValue,
	}
	validType := false
	switch interval.Unit {
	case Year:
		fallthrough
	case Month:
		fallthrough
	case Week:
		fallthrough
	case Day:
		fallthrough
	case Hour:
		validType = true
	}
	if !(validType && interval.Value >= 0) {
		interval = TimeInterval{
			Unit:  Month,
			Value: 1,
		}
	}

	if series.NextRecordingAfter.IsZero() {
		series.NextRecordingAfter = interval.StepForwards(now)
	}
	if series.NextSnapshotAfter.IsZero() {
		series.NextSnapshotAfter = nextSnapshot(now)
	}
	if series.OldestHistoricalAt.IsZero() {
		// TODO(insights): this value should probably somewhere more discoverable / obvious than here
		series.OldestHistoricalAt = now.Add(-time.Hour * 24 * 7 * 26)
	}

	seriesID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(`
			INSERT INTO insight_series (
				series_id,
				query,
				created_at,
				oldest_historical_at,
				last_recorded_at,
				next_recording_after,
				last_snapshot_at,
				next_snapshot_after
				repositories,
				sample_interval_unit,
				sample_interval_value,
				generated_from_capture_groups,
				just_in_time,
				generation_method,
				group_by,
				needs_migration,
			)
			VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, false)
			RETURNING id
		`,
		series.SeriesID,
		series.Query,
		series.CreatedAt,
		series.OldestHistoricalAt,
		series.LastRecordedAt,
		series.NextRecordingAfter,
		series.LastSnapshotAt,
		series.NextSnapshotAfter,
		pq.Array(series.Repositories),
		series.SampleIntervalUnit,
		series.SampleIntervalValue,
		series.GeneratedFromCaptureGroups,
		series.JustInTime,
		series.GenerationMethod,
		series.GroupBy,
	)))
	if err != nil {
		return errors.Wrapf(err, "unable to migrate insight series, unique_id: %s", from.ID)
	}
	series.ID = seriesID
	series.Enabled = true

	metadata := insightViewSeriesMetadata{}
	err = tx.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO insight_view_series (
			insight_series_id,
			insight_view_id,
			label,
			stroke
		)
		VALUES (%s, %s, %s, %s)
	`,
		series.ID,
		view.ID,
		metadata.Label,
		metadata.Stroke,
	))
	if err != nil {
		return err
	}
	// Enable the series in case it had previously been soft-deleted.
	err = tx.Exec(ctx, sqlf.Sprintf(`
		UPDATE insight_series
		SET deleted_at IS NULL
		WHERE series_id = %s
	`,
		series.SeriesID,
	))

	return nil
}

func migrateSeries(ctx context.Context, insightStore *basestore.Store, workerStore *basestore.Store, from searchInsight, batch migrationBatch) (err error) {
	tx, err := insightStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	dataSeries := make([]insightSeries, len(from.Series))
	metadata := make([]insightViewSeriesMetadata, len(from.Series))

	for i, timeSeries := range from.Series {
		temp := insightSeries{
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
			temp.GenerationMethod = Search
		} else if batch == backend {
			temp.SampleIntervalUnit = string(Month)
			temp.SampleIntervalValue = 1
			temp.NextRecordingAfter = nextRecording(time.Now())
			temp.NextSnapshotAfter = nextSnapshot(time.Now())
			temp.SeriesID = ksuid.New().String()
			temp.JustInTime = false
			temp.GenerationMethod = Search
		}

		var series insightSeries

		// Backend series require special consideration to re-use series
		if batch == backend {
			rows, err := scanSeries(tx.Query(ctx, sqlf.Sprintf(`
				SELECT
					id,
					series_id,
					query,
					created_at,
					oldest_historical_at,
					last_recorded_at,
					next_recording_after,
					last_snapshot_at,
					next_snapshot_after,
					(CASE WHEN deleted_at IS NULL THEN TRUE ELSE FALSE END) AS enabled,
					sample_interval_unit,
					sample_interval_value,
					generated_from_capture_groups,
					just_in_time,
					generation_method,
					repositories,
					group_by,
					backfill_attempts
				FROM insight_series
				WHERE
					(repositories = '{}' OR repositories is NULL) AND
					query = %s AND
					sample_interval_unit = %s AND
					sample_interval_value = %s AND
					generated_from_capture_groups = %s AND
					group_by IS NULL
			`,
				temp.Query,
				temp.SampleIntervalUnit,
				temp.SampleIntervalValue,
				false,
			)))
			if err != nil {
				return errors.Wrapf(err, "unable to migrate insight unique_id: %s series_id: %s", from.ID, temp.SeriesID)
			}
			if len(rows) > 0 {
				// If the series already exists, we can re-use that series
				series = rows[0]
			} else {
				now := time.Now()

				if temp.CreatedAt.IsZero() {
					temp.CreatedAt = now
				}
				interval := TimeInterval{
					Unit:  intervalUnit(temp.SampleIntervalUnit),
					Value: temp.SampleIntervalValue,
				}
				validType := false
				switch interval.Unit {
				case Year:
					fallthrough
				case Month:
					fallthrough
				case Week:
					fallthrough
				case Day:
					fallthrough
				case Hour:
					validType = true
				}
				if !(validType && interval.Value >= 0) {
					interval = TimeInterval{
						Unit:  Month,
						Value: 1,
					}
				}

				if temp.NextRecordingAfter.IsZero() {
					temp.NextRecordingAfter = interval.StepForwards(now)
				}
				if temp.NextSnapshotAfter.IsZero() {
					temp.NextSnapshotAfter = nextSnapshot(now)
				}
				if temp.OldestHistoricalAt.IsZero() {
					// TODO(insights): this value should probably somewhere more discoverable / obvious than here
					temp.OldestHistoricalAt = now.Add(-time.Hour * 24 * 7 * 26)
				}
				// If it's not a backend series, we just want to create it.
				id, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(`
				INSERT INTO insight_series (
					series_id,
					query,
					created_at,
					oldest_historical_at,
					last_recorded_at,
					next_recording_after,
					last_snapshot_at,
					next_snapshot_after,
					repositories,
					sample_interval_unit,
					sample_interval_value,
					generated_from_capture_groups,
					just_in_time,
					generation_method,
					group_by,
					needs_migration
					)
				VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, false)
				RETURNING id
			`,
					temp.SeriesID,
					temp.Query,
					temp.CreatedAt,
					temp.OldestHistoricalAt,
					temp.LastRecordedAt,
					temp.NextRecordingAfter,
					temp.LastSnapshotAt,
					temp.NextSnapshotAfter,
					pq.Array(temp.Repositories),
					temp.SampleIntervalUnit,
					temp.SampleIntervalValue,
					temp.GeneratedFromCaptureGroups,
					temp.JustInTime,
					temp.GenerationMethod,
					temp.GroupBy,
				)))
				if err != nil {
					return errors.Wrapf(err, "unable to migrate insight unique_id: %s series_id: %s", from.ID, temp.SeriesID)
				}
				temp.ID = id
				temp.Enabled = true
				series = temp

				// Also match/replace old series_points ids with the new series id
				oldId := fmt.Sprintf("s:%s", fmt.Sprintf("%X", sha256.Sum256([]byte(timeSeries.Query))))
				countUpdated, silentErr := updateTimeSeriesReferences(tx, ctx, oldId, temp.SeriesID)
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
						now := time.Now()
						silentErr := tx.Exec(ctx, sqlf.Sprintf(`UPDATE insight_series SET backfill_queued_at = %s WHERE id = %s`, now, series.ID))
						series.BackfillQueuedAt = now
						if silentErr != nil {
							// If the stamp fails, skip it. It will just need to be calcuated again.
							log15.Error("error updating backfill_queued_at", "series_id", temp.SeriesID, "err", silentErr)
						}
					}
				}
			}
		} else {
			now := time.Now()

			if temp.CreatedAt.IsZero() {
				temp.CreatedAt = now
			}
			interval := TimeInterval{
				Unit:  intervalUnit(temp.SampleIntervalUnit),
				Value: temp.SampleIntervalValue,
			}
			validType := false
			switch interval.Unit {
			case Year:
				fallthrough
			case Month:
				fallthrough
			case Week:
				fallthrough
			case Day:
				fallthrough
			case Hour:
				validType = true
			}
			if !(validType && interval.Value >= 0) {
				interval = TimeInterval{
					Unit:  Month,
					Value: 1,
				}
			}

			if temp.NextRecordingAfter.IsZero() {
				temp.NextRecordingAfter = interval.StepForwards(now)
			}
			if temp.NextSnapshotAfter.IsZero() {
				temp.NextSnapshotAfter = nextSnapshot(now)
			}
			if temp.OldestHistoricalAt.IsZero() {
				// TODO(insights): this value should probably somewhere more discoverable / obvious than here
				temp.OldestHistoricalAt = now.Add(-time.Hour * 24 * 7 * 26)
			}
			// If it's not a backend series, we just want to create it.
			id, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(`
				INSERT INTO insight_series (
					series_id,
					query,
					created_at,
					oldest_historical_at,
					last_recorded_at,
					next_recording_after,
					last_snapshot_at,
					next_snapshot_after,
					repositories,
					sample_interval_unit,
					sample_interval_value,
					generated_from_capture_groups,
					just_in_time,
					generation_method,
					group_by,
					needs_migration
					)
				VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, false)
				RETURNING id
			`,
				temp.SeriesID,
				temp.Query,
				temp.CreatedAt,
				temp.OldestHistoricalAt,
				temp.LastRecordedAt,
				temp.NextRecordingAfter,
				temp.LastSnapshotAt,
				temp.NextSnapshotAfter,
				pq.Array(temp.Repositories),
				temp.SampleIntervalUnit,
				temp.SampleIntervalValue,
				temp.GeneratedFromCaptureGroups,
				temp.JustInTime,
				temp.GenerationMethod,
				temp.GroupBy,
			)))
			if err != nil {
				return errors.Wrapf(err, "unable to migrate insight unique_id: %s series_id: %s", from.ID, temp.SeriesID)
			}
			temp.ID = id
			temp.Enabled = true
			series = temp
		}
		dataSeries[i] = series

		metadata[i] = insightViewSeriesMetadata{
			Label:  timeSeries.Name,
			Stroke: timeSeries.Stroke,
		}
	}

	view := insightView{
		Title:            from.Title,
		Description:      from.Description,
		UniqueID:         from.ID,
		PresentationType: Line,
	}

	if from.Filters != nil {
		view.Filters = insightViewFilters{
			IncludeRepoRegex: from.Filters.IncludeRepoRegexp,
			ExcludeRepoRegex: from.Filters.ExcludeRepoRegexp,
		}
	}

	var grants []insightViewGrant
	if from.UserID != nil {
		grants = []insightViewGrant{userGrant(int(*from.UserID))}
	} else if from.OrgID != nil {
		grants = []insightViewGrant{orgGrant(int(*from.OrgID))}
	} else {
		grants = []insightViewGrant{globalGrant()}
	}

	viewID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(`
		INSERT INTO insight_view (
			title,
			description,
			unique_id,
			default_filter_include_repo_regex,
			default_filter_exclude_repo_regex,
			default_filter_search_contexts,
			other_threshold,
			presentation_type,
		)
		VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
		RETURNING id
	`,
		view.Title,
		view.Description,
		view.UniqueID,
		view.Filters.IncludeRepoRegex,
		view.Filters.ExcludeRepoRegex,
		pq.Array(view.Filters.SearchContexts),
		view.OtherThreshold,
		view.PresentationType,
	)))
	view.ID = viewID

	values := make([]*sqlf.Query, 0, len(grants))
	for _, grant := range grants {
		values = append(values, sqlf.Sprintf("(%s, %s, %s, %s)", view.ID, grant.OrgID, grant.UserID, grant.Global))
	}
	err = tx.Exec(ctx, sqlf.Sprintf(`INSERT INTO insight_view_grants (insight_view_id, org_id, user_id, global) VALUES %s`, sqlf.Join(values, ", ")))
	if err != nil {
		return err
	}

	for i, insightSeries := range dataSeries {
		err = tx.Exec(ctx, sqlf.Sprintf(
			`INSERT INTO insight_view_series (
				insight_series_id,
				insight_view_id,
				label,
				stroke,
			)
			VALUES (%s, %s, %s, %s)
		`,
			insightSeries.ID,
			view.ID,
			metadata[i].Label,
			metadata[i].Stroke,
		))
		if err != nil {
			return err
		}

		err = tx.Exec(ctx, sqlf.Sprintf(`UPDATE insight_series SET deleted_at IS NULL WHERE series_id = %s`, insightSeries.SeriesID))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *migrator) migrateDashboards(ctx context.Context, toMigrate []settingDashboard, mc migrationContext) (int, error) {
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
func parseTimeInterval(insight searchInsight) timeInterval {
	if insight.Step.Days != nil {
		return timeInterval{
			unit:  Day,
			value: *insight.Step.Days,
		}
	} else if insight.Step.Hours != nil {
		return timeInterval{
			unit:  Hour,
			value: *insight.Step.Hours,
		}
	} else if insight.Step.Weeks != nil {
		return timeInterval{
			unit:  Week,
			value: *insight.Step.Weeks,
		}
	} else if insight.Step.Months != nil {
		return timeInterval{
			unit:  Month,
			value: *insight.Step.Months,
		}
	} else if insight.Step.Years != nil {
		return timeInterval{
			unit:  Year,
			value: *insight.Step.Years,
		}
	} else {
		return timeInterval{
			unit:  Month,
			value: 1,
		}
	}
}

func makeUniqueId(id string, subject settingsSubject) string {
	if subject.User != nil {
		return fmt.Sprintf("%s-user-%d", id, *subject.User)
	} else if subject.Org != nil {
		return fmt.Sprintf("%s-org-%d", id, *subject.Org)
	} else {
		return id
	}
}

func getOwnerName(settingsRow settings) string {
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

func getOwnerNameFromInsight(insight searchInsight) string {
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

func getOwnerNameFromLangStatsInsight(insight langStatsInsight) string {
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

func getOwnerNameFromDashboard(insight settingDashboard) string {
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
