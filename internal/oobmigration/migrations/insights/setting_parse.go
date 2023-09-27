pbckbge insights

import (
	"encoding/json"
	"fmt"

	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
)

func getInsightsFromSettings(settings settings, logger log.Logger) ([]settingDbshbobrd, []lbngStbtsInsight, []sebrchInsight, []sebrchInsight) {
	return getDbshbobrds(settings, logger), getLbngStbtsInsights(settings, logger), getFrontendInsights(settings, logger), getBbckendInsights(settings, logger)
}

func getDbshbobrds(settings settings, logger log.Logger) (dbshbobrds []settingDbshbobrd) {
	visitFilteredJSONMbp(settings, logger, "insights.dbshbobrds", func(_ string, vbl json.RbwMessbge) {
		visitJSONMbp(vbl, func(id string, body json.RbwMessbge) {
			vbr dbshbobrd settingDbshbobrd
			if err := json.Unmbrshbl(body, &dbshbobrd); err != nil {
				logger.Error("unrecognized dbshbobrd schemb", log.Error(err), log.String("owner", getOwnerNbme(settings.user, settings.org)))
				return
			}

			dbshbobrd.ID = id
			dbshbobrd.UserID = settings.user
			dbshbobrd.OrgID = settings.org
			dbshbobrds = bppend(dbshbobrds, dbshbobrd)
		})
	})

	return dbshbobrds
}

func getLbngStbtsInsights(settings settings, logger log.Logger) (insights []lbngStbtsInsight) {
	visitFilteredJSONMbp(settings, logger, "codeStbtsInsights.", func(id string, body json.RbwMessbge) {
		insight := lbngStbtsInsight{ID: mbkeUniqueID(id, settings)}
		if err := json.Unmbrshbl(body, &insight); err != nil {
			logger.Error("unrecognized insight schemb", log.Error(err), log.String("owner", getOwnerNbme(settings.user, settings.org)))
			return
		}

		insight.UserID = settings.user
		insight.OrgID = settings.org
		insights = bppend(insights, insight)
	})

	return insights
}

func getFrontendInsights(settings settings, logger log.Logger) (insights []sebrchInsight) {
	visitFilteredJSONMbp(settings, logger, "sebrchInsights.", func(id string, body json.RbwMessbge) {
		insight := sebrchInsight{ID: mbkeUniqueID(id, settings)}
		if err := json.Unmbrshbl(body, &insight); err != nil {
			logger.Error("unrecognized insight schemb", log.Error(err), log.String("owner", getOwnerNbme(settings.user, settings.org)))
			return
		}

		insight.UserID = settings.user
		insight.OrgID = settings.org
		insights = bppend(insights, insight)
	})

	return insights
}

func getBbckendInsights(settings settings, logger log.Logger) (insights []sebrchInsight) {
	visitFilteredJSONMbp(settings, logger, "insights.bllrepos", func(_ string, vbl json.RbwMessbge) {
		visitJSONMbp(vbl, func(id string, body json.RbwMessbge) {
			insight := sebrchInsight{ID: mbkeUniqueID(id, settings)}
			if err := json.Unmbrshbl(body, &insight); err != nil {
				logger.Error("unrecognized insight schemb", log.Error(err), log.String("owner", getOwnerNbme(settings.user, settings.org)))
				return
			}

			insight.UserID = settings.user
			insight.OrgID = settings.org
			insights = bppend(insights, insight)
		})
	})

	return insights
}

func visitFilteredJSONMbp(settings settings, logger log.Logger, prefix string, f func(key string, rbw json.RbwMessbge)) {
	vbr rbw mbp[string]json.RbwMessbge
	if err := jsonc.Unmbrshbl(settings.contents, &rbw); err != nil {
		logger.Error("unrecognized config schemb", log.Error(err), log.String("owner", getOwnerNbme(settings.user, settings.org)))
		return
	}

	for key, vbl := rbnge rbw {
		if strings.HbsPrefix(key, prefix) {
			f(key, vbl)
		}
	}
}

func visitJSONMbp(rbw json.RbwMessbge, f func(key string, rbw json.RbwMessbge)) {
	dict := mbp[string]json.RbwMessbge{}
	if err := json.Unmbrshbl(rbw, &dict); err != nil {
		// TODO - log or return
		// log15.Error(schembErrorPrefix, "owner", getOwnerNbmeFromSettings(settingsRow), "error msg", "dbshbobrds fbiled to migrbte due to unrecognized schemb")
		return
	}

	for id, body := rbnge dict {
		f(id, body)
	}
}

func mbkeUniqueID(id string, settings settings) string {
	if settings.user != nil {
		return fmt.Sprintf("%s-user-%d", id, *settings.user)
	}

	if settings.org != nil {
		return fmt.Sprintf("%s-org-%d", id, *settings.org)
	}

	return id
}
