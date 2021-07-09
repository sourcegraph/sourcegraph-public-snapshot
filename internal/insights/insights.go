package insights

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
)

// GetSettings returns all settings on the Sourcegraph installation that can be filtered by a type. This is useful for
// generating aggregates for code insights which are currently stored in the settings.
// 🚨 SECURITY: This method bypasses any user permissions to fetch a list of all settings on the Sourcegraph installation.
//It is used for generating aggregated analytics that require an accurate view across all settings, such as for code insights🚨
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

func GetSearchInsights(ctx context.Context, db dbutil.DB, filter SettingFilter) ([]SearchInsight, error) {
	prefix := "searchInsights."
	settings, err := GetSettings(ctx, db, filter, prefix)
	if err != nil {
		return []SearchInsight{}, err
	}

	var raw map[string]json.RawMessage
	results := make([]SearchInsight, 0)

	for _, setting := range settings {
		raw, err = FilterSettingJson(setting.Contents, prefix)
		if err != nil {
			return []SearchInsight{}, err
		}

		var temp SearchInsight

		for id, body := range raw {
			temp.ID = id
			if err := json.Unmarshal(body, &temp); err != nil {
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

	for _, setting := range settings {
		raw, err = FilterSettingJson(setting.Contents, prefix)
		if err != nil {
			return []LangStatsInsight{}, err
		}

		var temp LangStatsInsight

		for id, body := range raw {
			temp.ID = id
			if err := json.Unmarshal(body, &temp); err != nil {
				// a deprecated schema collides with this field name, so skip any deserialization errors
				continue
			}
			results = append(results, temp)
		}
	}
	return results, nil
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
}

type LangStatsInsight struct {
	ID             string
	Title          string
	Repository     string
	OtherThreshold float32
}

type SettingFilter string

const (
	Org  SettingFilter = "org"
	User SettingFilter = "user"
	All  SettingFilter = "all"
)
