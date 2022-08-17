package insights

import (
	"encoding/json"
	"fmt"

	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"
)

func getInsightsFromSettings(settings settings, logger log.Logger) ([]settingDashboard, []langStatsInsight, []searchInsight, []searchInsight) {
	return getDashboards(settings, logger), getLangStatsInsights(settings, logger), getFrontendInsights(settings, logger), getBackendInsights(settings, logger)
}

func getDashboards(settings settings, logger log.Logger) (dashboards []settingDashboard) {
	visitFilteredJSONMap(settings, logger, "insights.dashboards", func(_ string, val json.RawMessage) {
		visitJSONMap(val, func(id string, body json.RawMessage) {
			var dashboard settingDashboard
			if err := json.Unmarshal(body, &dashboard); err != nil {
				logger.Error("unrecognized dashboard schema", log.Error(err), log.String("owner", getOwnerName(settings.user, settings.org)))
				return
			}

			dashboard.ID = id
			dashboard.UserID = settings.user
			dashboard.OrgID = settings.org
			dashboards = append(dashboards, dashboard)
		})
	})

	return dashboards
}

func getLangStatsInsights(settings settings, logger log.Logger) (insights []langStatsInsight) {
	visitFilteredJSONMap(settings, logger, "codeStatsInsights.", func(id string, body json.RawMessage) {
		insight := langStatsInsight{ID: makeUniqueID(id, settings)}
		if err := json.Unmarshal(body, &insight); err != nil {
			logger.Error("unrecognized insight schema", log.Error(err), log.String("owner", getOwnerName(settings.user, settings.org)))
			return
		}

		insight.UserID = settings.user
		insight.OrgID = settings.org
		insights = append(insights, insight)
	})

	return insights
}

func getFrontendInsights(settings settings, logger log.Logger) (insights []searchInsight) {
	visitFilteredJSONMap(settings, logger, "searchInsights.", func(id string, body json.RawMessage) {
		insight := searchInsight{ID: makeUniqueID(id, settings)}
		if err := json.Unmarshal(body, &insight); err != nil {
			logger.Error("unrecognized insight schema", log.Error(err), log.String("owner", getOwnerName(settings.user, settings.org)))
			return
		}

		insight.UserID = settings.user
		insight.OrgID = settings.org
		insights = append(insights, insight)
	})

	return insights
}

func getBackendInsights(settings settings, logger log.Logger) (insights []searchInsight) {
	visitFilteredJSONMap(settings, logger, "insights.allrepos", func(_ string, val json.RawMessage) {
		visitJSONMap(val, func(id string, body json.RawMessage) {
			insight := searchInsight{ID: makeUniqueID(id, settings)}
			if err := json.Unmarshal(body, &insight); err != nil {
				logger.Error("unrecognized insight schema", log.Error(err), log.String("owner", getOwnerName(settings.user, settings.org)))
				return
			}

			insight.UserID = settings.user
			insight.OrgID = settings.org
			insights = append(insights, insight)
		})
	})

	return insights
}

func visitFilteredJSONMap(settings settings, logger log.Logger, prefix string, f func(key string, raw json.RawMessage)) {
	var raw map[string]json.RawMessage
	if err := jsonc.Unmarshal(settings.contents, &raw); err != nil {
		logger.Error("unrecognized config schema", log.Error(err), log.String("owner", getOwnerName(settings.user, settings.org)))
		return
	}

	for key, val := range raw {
		if strings.HasPrefix(key, prefix) {
			f(key, val)
		}
	}
}

func visitJSONMap(raw json.RawMessage, f func(key string, raw json.RawMessage)) {
	dict := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &dict); err != nil {
		// TODO - log or return
		// log15.Error(schemaErrorPrefix, "owner", getOwnerNameFromSettings(settingsRow), "error msg", "dashboards failed to migrate due to unrecognized schema")
		return
	}

	for id, body := range dict {
		f(id, body)
	}
}

func makeUniqueID(id string, settings settings) string {
	if settings.user != nil {
		return fmt.Sprintf("%s-user-%d", id, *settings.user)
	}

	if settings.org != nil {
		return fmt.Sprintf("%s-org-%d", id, *settings.org)
	}

	return id
}
