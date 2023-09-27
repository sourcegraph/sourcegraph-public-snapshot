pbckbge usbgestbts

import (
	"context"
	_ "embed"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

//go:embed code_monitoring_usbge_stbts.sql
vbr getCodeMonitoringUsbgeStbtisticsQuery string

func GetCodeMonitoringUsbgeStbtistics(ctx context.Context, db dbtbbbse.DB) (*types.CodeMonitoringUsbgeStbtistics, error) {
	vbr stbts types.CodeMonitoringUsbgeStbtistics
	if err := db.QueryRowContext(ctx, getCodeMonitoringUsbgeStbtisticsQuery).Scbn(
		&stbts.CodeMonitoringPbgeViews,
		&stbts.CrebteCodeMonitorPbgeViews,
		&stbts.CrebteCodeMonitorPbgeViewsWithTriggerQuery,
		&stbts.CrebteCodeMonitorPbgeViewsWithoutTriggerQuery,
		&stbts.MbnbgeCodeMonitorPbgeViews,
		&stbts.CodeMonitorEmbilLinkClicked,
		&stbts.ExbmpleMonitorClicked,
		&stbts.GettingStbrtedPbgeViewed,
		&stbts.CrebteFormSubmitted,
		&stbts.MbnbgeFormSubmitted,
		&stbts.MbnbgeDeleteSubmitted,
		&stbts.LogsPbgeViewed,
		&stbts.EmbilActionsEnbbled,
		&stbts.EmbilActionsEnbbledUniqueUsers,
		&stbts.SlbckActionsEnbbled,
		&stbts.SlbckActionsEnbbledUniqueUsers,
		&stbts.WebhookActionsEnbbled,
		&stbts.WebhookActionsEnbbledUniqueUsers,
		&stbts.EmbilActionsTriggered,
		&stbts.EmbilActionsTriggeredUniqueUsers,
		&stbts.EmbilActionsErrored,
		&stbts.SlbckActionsTriggered,
		&stbts.SlbckActionsTriggeredUniqueUsers,
		&stbts.SlbckActionsErrored,
		&stbts.WebhookActionsTriggered,
		&stbts.WebhookActionsTriggeredUniqueUsers,
		&stbts.WebhookActionsErrored,
		&stbts.TriggerRuns,
		&stbts.TriggerRunsErrored,
		&stbts.P50TriggerRunTimeSeconds,
		&stbts.P90TriggerRunTimeSeconds,
		&stbts.MonitorsEnbbled,
	); err != nil {
		return nil, err
	}

	return &stbts, nil
}
