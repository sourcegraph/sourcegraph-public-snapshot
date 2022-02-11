package insights

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Loader will load insights from some persistent storage.
type Loader interface {
	LoadAll(ctx context.Context) ([]SearchInsight, error)
	LoadDashboards(ctx context.Context) ([]SettingDashboard, error)
}

// DBLoader will load insights from a database. This is also where the application can access insights currently stored
// in user / org settings.
type DBLoader struct {
	db dbutil.DB
}

func (d *DBLoader) LoadAll(ctx context.Context) ([]SearchInsight, error) {
	return GetIntegratedInsights(ctx, d.db)
}

func (d *DBLoader) LoadDashboards(ctx context.Context) ([]SettingDashboard, error) {
	return DiscoverDashboardsInSettings(ctx, d.db)
}

func NewLoader(db dbutil.DB) Loader {
	return &DBLoader{db: db}
}

// GetSettings returns all settings on the Sourcegraph installation that can be filtered by a type. This is useful for
// generating aggregates for code insights which are currently stored in the settings.
// 🚨 SECURITY: This method bypasses any user permissions to fetch a list of all settings on the Sourcegraph installation.
// It is used for generating aggregated analytics that require an accurate view across all settings, such as for code insights🚨
func GetSettings(ctx context.Context, db dbutil.DB, filter SettingFilter, prefix string) ([]*api.Settings, error) {
	settingStore := database.Settings(db)
	settings, err := settingStore.ListAll(ctx, prefix)
	if err != nil {
		return []*api.Settings{}, err
	}
	filtered := make([]*api.Settings, 0)

	for _, setting := range settings {
		if setting.Subject.Org != nil && filter == Org {
			filtered = append(filtered, setting)
		} else if setting.Subject.User != nil && filter == User {
			filtered = append(filtered, setting)
		} else if filter == All {
			filtered = append(filtered, setting)
		}
	}

	return filtered, nil
}

// FilterSettingJson will return a json map that only contains keys that match a prefix string, mapped to the keyed contents.
func FilterSettingJson(settingJson string, prefix string) (map[string]json.RawMessage, error) {
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

// GetSearchInsights returns insights stored in user / org / global settings that match the extensions schema. This schema is planned for deprecation
// and currently only exists to service pings.
func GetSearchInsights(ctx context.Context, db dbutil.DB, filter SettingFilter) ([]SearchInsight, error) {
	prefix := "searchInsights."
	settings, err := GetSettings(ctx, db, filter, prefix)
	if err != nil {
		return []SearchInsight{}, err
	}

	var raw map[string]json.RawMessage
	results := make([]SearchInsight, 0)

	for i := range settings {
		raw, err = FilterSettingJson(settings[i].Contents, prefix)
		if err != nil {
			return []SearchInsight{}, err
		}

		for id := range raw {
			var temp SearchInsight
			temp.ID = id
			if err := json.Unmarshal(raw[id], &temp); err != nil {
				// a deprecated schema collides with this field name, so skip any deserialization errors
				continue
			}
			results = append(results, temp)
		}
	}
	return results, nil
}

func GetLangStatsInsights(ctx context.Context, db dbutil.DB, filter SettingFilter) ([]LangStatsInsight, error) {
	prefix := "codeStatsInsights."

	settings, err := GetSettings(ctx, db, filter, prefix)
	if err != nil {
		return []LangStatsInsight{}, err
	}

	var raw map[string]json.RawMessage
	results := make([]LangStatsInsight, 0)

	for i := range settings {
		raw, err = FilterSettingJson(settings[i].Contents, prefix)
		if err != nil {
			return []LangStatsInsight{}, err
		}
		userId := settings[i].Subject.User
		orgId := settings[i].Subject.Org

		for id, body := range raw {
			var temp LangStatsInsight
			temp.ID = id
			if err := json.Unmarshal(body, &temp); err != nil {
				// a deprecated schema collides with this field name, so skip any deserialization errors
				continue
			}
			temp.UserID = userId
			temp.OrgID = orgId

			results = append(results, temp)
		}
	}
	return results, nil
}

// GetIntegratedInsights returns all of the insights defined by the extension based Code Insights that are compatible
// running over all repositories. These are located in a specific setting object `insights.allrepos` which is a
// dictionary of unique keys to extension setting body. This is intended to be deprecated as soon as code insights migrates
// fully to a persistent database. Any deserialization errors that occur during parsing will be logged as errors, but will not
// cause any errors to surface.
func GetIntegratedInsights(ctx context.Context, db dbutil.DB) ([]SearchInsight, error) {
	prefix := "insights.allrepos"

	settings, err := GetSettings(ctx, db, All, prefix)
	if err != nil {
		return []SearchInsight{}, err
	}

	var multi error

	results := make([]SearchInsight, 0)
	for i := range settings {
		setting := settings[i]
		perms := permissionAssociations{
			userID: setting.Subject.User,
			orgID:  setting.Subject.Org,
		}

		var raw map[string]json.RawMessage
		raw, err = FilterSettingJson(setting.Contents, prefix)
		if err != nil {
			multi = errors.Append(multi, err)
			continue
		}

		for _, val := range raw {
			// iterate for each instance of the prefix key in the settings. This should never be len > 1, but it's technically a map.
			temp, err := unmarshalIntegrated(val)
			if err != nil {
				// this isn't actually a total failure case, we could have partially parsed this dictionary.
				multi = errors.Append(multi, err)
			}
			results = append(results, temp.Insights(perms)...)
		}
	}

	if multi != nil {
		log15.Error("insights: deserialization errors parsing integrated insights", "error", multi)
	}

	return results, nil
}

// IntegratedInsights represents a settings dictionary of valid insights that are integrated across the extensions API and the backend.
type IntegratedInsights map[string]SearchInsight

// unmarshalIntegrated will attempt to unmarshall a JSON dictionary where each key represents a unique id and each value represents a SearchInsight.
// Errors will be collected and reported out, but will not fail the entire unmarshal if possible.
func unmarshalIntegrated(raw json.RawMessage) (IntegratedInsights, error) {
	var dict map[string]json.RawMessage
	var multi error
	result := make(IntegratedInsights)

	if err := json.Unmarshal(raw, &dict); err != nil {
		return result, err
	}

	for id, body := range dict {
		var temp SearchInsight
		if err := json.Unmarshal(body, &temp); err != nil {
			multi = errors.Append(multi, err)
			continue
		}
		result[id] = temp
	}

	return result, multi
}

// permissionAssociations contains user / org information that is derived from a setting
type permissionAssociations struct {
	userID *int32
	orgID  *int32
}

// Insights returns an array of contained insights.
func (i IntegratedInsights) Insights(perms permissionAssociations) []SearchInsight {
	results := make([]SearchInsight, 0)
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

type TimeSeries struct {
	Name   string
	Stroke string
	Query  string
}

type Interval struct {
	Years  *int
	Months *int
	Weeks  *int
	Days   *int
	Hours  *int
}

type SearchInsight struct {
	ID           string
	Title        string
	Description  string
	Repositories []string
	Series       []TimeSeries
	Step         Interval
	Visibility   string
	OrgID        *int32
	UserID       *int32
	Filters      *DefaultFilters
}

type LangStatsInsight struct {
	ID             string
	Title          string
	Repository     string
	OtherThreshold float64
	OrgID          *int32
	UserID         *int32
}

type DefaultFilters struct {
	IncludeRepoRegexp *string
	ExcludeRepoRegexp *string
}

type SettingFilter string

const (
	Org  SettingFilter = "org"
	User SettingFilter = "user"
	All  SettingFilter = "all"
)

type SettingDashboard struct {
	ID         string   `json:"id,omitempty"`
	Title      string   `json:"title,omitempty"`
	InsightIds []string `json:"insightIds,omitempty"`
	UserID     *int32
	OrgID      *int32
}

func DiscoverDashboardsInSettings(ctx context.Context, db dbutil.DB) ([]SettingDashboard, error) {
	prefix := "insights.dashboards"
	settings, err := GetSettings(ctx, db, All, prefix)
	if err != nil {
		return []SettingDashboard{}, err
	}
	var multi error

	results := make([]SettingDashboard, 0)
	for _, setting := range settings {
		perms := permissionAssociations{
			userID: setting.Subject.User,
			orgID:  setting.Subject.Org,
		}

		var raw map[string]json.RawMessage
		raw, err := FilterSettingJson(setting.Contents, prefix)
		if err != nil {
			multi = errors.Append(multi, err)
			continue
		}
		for _, val := range raw {
			// iterate for each instance of the prefix key in the settings. This should never be len > 1, but it's technically a map.
			temp, err := unmarshalDashboard(val)
			if err != nil {
				// this isn't actually a total failure case, we could have partially parsed this dictionary.
				multi = errors.Append(multi, err)
			}
			results = append(results, temp.Dashboards(perms)...)
		}
	}
	if multi != nil {
		log15.Error("insights: deserialization errors parsing integrated dashboards", "error", multi)
	}
	return results, nil
}

// Dashboards returns an array of contained dashboards.
func (i IntegratedDashboards) Dashboards(perms permissionAssociations) []SettingDashboard {
	results := make([]SettingDashboard, 0, len(i))
	for key, dashboard := range i {
		dashboard.ID = key // the insight ID is the value of the dict key

		// each setting is owned by either a user or an organization, which needs to be mapped when this insight is synced
		// to preserve permissions semantics
		dashboard.UserID = perms.userID
		dashboard.OrgID = perms.orgID

		results = append(results, dashboard)
	}
	return results
}

type IntegratedDashboards map[string]SettingDashboard

// unmarshalIntegrated will attempt to unmarshall a JSON dictionary where each key represents a unique id and each value represents a SearchInsight.
// Errors will be collected and reported out, but will not fail the entire unmarshal if possible.
func unmarshalDashboard(raw json.RawMessage) (IntegratedDashboards, error) {
	var dict map[string]json.RawMessage
	var multi error
	result := make(IntegratedDashboards)

	if err := json.Unmarshal(raw, &dict); err != nil {
		return result, err
	}

	for id, body := range dict {
		var temp SettingDashboard
		if err := json.Unmarshal(body, &temp); err != nil {
			multi = errors.Append(multi, err)
			continue
		}
		result[id] = temp
	}

	return result, multi
}

// NextRecording calculates the time that a series recording should occur given the current or most recent recording time.
func NextRecording(current time.Time) time.Time {
	year, month, _ := current.In(time.UTC).Date()
	return time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
}

func NextSnapshot(current time.Time) time.Time {
	year, month, day := current.In(time.UTC).Date()
	return time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)
}
