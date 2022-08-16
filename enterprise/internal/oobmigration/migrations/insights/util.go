package insights

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"
)

// replaceIfEmpty will return a string where the first argument is given priority if non-empty.
func replaceIfEmpty(firstChoice *string, replacement string) string {
	if firstChoice == nil || *firstChoice == "" {
		return replacement
	}
	return *firstChoice
}

func specialCaseDashboardTitle(subjectName string) string {
	format := "%s Insights"
	if subjectName == "Global" {
		return fmt.Sprintf(format, subjectName)
	}
	return fmt.Sprintf(format, fmt.Sprintf("%s's", subjectName))
}

func logDuplicates(insightIds []string) {
	set := make(map[string]struct{}, len(insightIds))
	for _, id := range insightIds {
		if _, ok := set[id]; ok {
			log15.Info("insights setting oob-migration: duplicate insight ID", "uniqueId", id)
		} else {
			set[id] = struct{}{}
		}
	}
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

// nextRecording calculates the time that a series recording should occur given the current or most recent recording time.
func nextRecording(current time.Time) time.Time {
	year, month, _ := current.In(time.UTC).Date()
	return time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
}

func nextSnapshot(current time.Time) time.Time {
	year, month, day := current.In(time.UTC).Date()
	return time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)
}

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
