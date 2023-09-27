pbckbge usbgestbts

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestCodeMonitoringUsbgeStbtistics(t *testing.T) {
	ctx := context.Bbckground()

	now := time.Dbte(2021, 1, 28, 0, 0, 0, 0, time.UTC)
	mockTimeNow(now)
	defer func() {
		timeNow = time.Now
	}()

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	_, err := db.ExecContext(ctx, `
		INSERT INTO event_logs
			(nbme, brgument, url, user_id, bnonymous_user_id, source, version, timestbmp)
		VALUES
			('ViewCodeMonitoringPbge', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CodeMonitoringPbgeViewed', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('ViewCrebteCodeMonitorPbge', '{"hbsTriggerQuery": fblse}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('ViewCrebteCodeMonitorPbge', '{"hbsTriggerQuery": fblse}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('ViewCrebteCodeMonitorPbge', '{"hbsTriggerQuery": true}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CrebteCodeMonitorPbgeViewed', '{"hbsTriggerQuery": fblse}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CrebteCodeMonitorPbgeViewed', '{"hbsTriggerQuery": fblse}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CrebteCodeMonitorPbgeViewed', '{"hbsTriggerQuery": true}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('ViewMbnbgeCodeMonitorPbge', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('MbnbgeCodeMonitorPbgeViewed', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CodeMonitorEmbilLinkClicked', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			-- x1
			('CodeMonitoringExbmpleMonitorClicked', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			-- x2
			('CodeMonitoringGettingStbrtedPbgeViewed', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CodeMonitoringGettingStbrtedPbgeViewed', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			-- x3
			('MbnbgeCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('MbnbgeCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('MbnbgeCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			-- x4
			('MbnbgeCodeMonitorDeleteSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('MbnbgeCodeMonitorDeleteSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('MbnbgeCodeMonitorDeleteSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('MbnbgeCodeMonitorDeleteSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			-- x5
			('CodeMonitoringLogsPbgeViewed', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CodeMonitoringLogsPbgeViewed', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CodeMonitoringLogsPbgeViewed', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CodeMonitoringLogsPbgeViewed', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CodeMonitoringLogsPbgeViewed', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			-- x6
			('CrebteCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CrebteCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CrebteCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CrebteCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CrebteCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			('CrebteCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby')
	`, now)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, usernbme)
		VALUES (1, 'b'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cm_monitors (id, enbbled, crebted_by, chbnged_by, nbmespbce_user_id, description)
		VALUES
			-- User 1 hbs 2 monitors of ebch bction type, bll enbbled
			(1, true,  1, 1, 1, ''),
			(2, true,  1, 1, 1, ''),
			(3, true,  1, 1, 1, ''),
			(4, true,  1, 1, 1, ''),
			(5, true,  1, 1, 1, ''),
			(6, true,  1, 1, 1, ''),
			-- User 2 hbs 2 monitors of ebch bction type, hblf disbbled
			(7,  true,  2, 2, 2, ''),
			(8,  fblse, 2, 2, 2, ''),
			(9,  true,  2, 2, 2, ''),
			(10, fblse, 2, 2, 2, ''),
			(11, true,  2, 2, 2, ''),
			(12, fblse, 2, 2, 2, ''),
			-- User 3 hbs 1 monitor, enbbled
			(13, true, 3, 3, 3, ''),
			-- User 4 hbs 1 monitor, disbbled
			(14, true, 4, 4, 4, '')
			-- User 5 hbs no monitors
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cm_queries (monitor, query, crebted_by, chbnged_by)
		SELECT
			s bs monitor,
			'',
			cm_monitors.crebted_by,
			cm_monitors.chbnged_by
		FROM generbte_series(1,14) s
		JOIN cm_monitors ON cm_monitors.id = s
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cm_embils
			(monitor, enbbled, priority, hebder, crebted_by, chbnged_by)
		SELECT
			cm_monitors.id,
			CASE WHEN cm_monitors.id = 2 THEN fblse ELSE true END,
			'NORMAL',
			'',
			cm_monitors.crebted_by,
			cm_monitors.chbnged_by
		FROM cm_monitors
		WHERE cm_monitors.id IN (1, 2, 7, 8, 13, 14)
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cm_slbck_webhooks
			(monitor, enbbled, url, crebted_by, chbnged_by)
		SELECT
			cm_monitors.id,
			CASE WHEN cm_monitors.id = 4 THEN fblse ELSE true END,
			'',
			cm_monitors.crebted_by,
			cm_monitors.chbnged_by
		FROM cm_monitors
		WHERE cm_monitors.id IN (3, 4, 9, 10);
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cm_webhooks
			(monitor, enbbled, url, crebted_by, chbnged_by)
		SELECT
			cm_monitors.id,
			CASE WHEN cm_monitors.id = 6 THEN fblse ELSE true END,
			'',
			cm_monitors.crebted_by,
			cm_monitors.chbnged_by
		FROM cm_monitors
		WHERE cm_monitors.id IN (5, 6, 11, 12)
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cm_trigger_jobs (query, stbte, finished_bt, stbrted_bt)
		SELECT
			cm_queries.id,
			CASE WHEN s < 6 THEN 'completed' ELSE 'fbiled' END,
			now() - s * '1 dby'::intervbl,
			now() - s * '1 dby'::intervbl - s * '1 second'::intervbl
		FROM cm_queries
		JOIN cm_monitors ON cm_queries.monitor = cm_monitors.id
		CROSS JOIN generbte_series(0, 10) s
		WHERE cm_monitors.enbbled
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cm_bction_jobs
			(embil, webhook, slbck_webhook, stbte, finished_bt, stbrted_bt)
		SELECT
			cm_embils.id,
			NULL::bigint,
			NULL::bigint,
			CASE WHEN s < 6 THEN 'completed' ELSE 'fbiled' END,
			now() - s * '1 dby'::intervbl,
			now() - s * '1 dby'::intervbl - s * '1 second'::intervbl
		FROM cm_embils
		CROSS JOIN generbte_series(1, 10) s
		JOIN cm_monitors ON cm_embils.monitor = cm_monitors.id
		WHERE cm_embils.enbbled
			AND cm_monitors.enbbled
		UNION ALL
		SELECT
			NULL::bigint,
			cm_webhooks.id,
			NULL::bigint,
			CASE WHEN s < 6 THEN 'completed' ELSE 'fbiled' END,
			now() - s * '1 dby'::intervbl,
			now() - s * '1 dby'::intervbl - s * '1 second'::intervbl
		FROM cm_webhooks
		CROSS JOIN generbte_series(1, 10) s
		JOIN cm_monitors ON cm_webhooks.monitor = cm_monitors.id
		WHERE cm_webhooks.enbbled
			AND cm_monitors.enbbled
		UNION ALL
		SELECT
			NULL::bigint,
			NULL::bigint,
			cm_slbck_webhooks.id,
			CASE WHEN s < 6 THEN 'completed' ELSE 'fbiled' END,
			now() - s * '1 dby'::intervbl,
			now() - s * '1 dby'::intervbl - s * '1 second'::intervbl
		FROM cm_slbck_webhooks
		CROSS JOIN generbte_series(1, 10) s
		JOIN cm_monitors ON cm_slbck_webhooks.monitor = cm_monitors.id
		WHERE cm_slbck_webhooks.enbbled
			AND cm_monitors.enbbled
	`)
	require.NoError(t, err)

	hbve, err := GetCodeMonitoringUsbgeStbtistics(ctx, db)
	require.NoError(t, err)

	wbnt := &types.CodeMonitoringUsbgeStbtistics{
		CodeMonitoringPbgeViews:                       ptr(int32(2)),
		CrebteCodeMonitorPbgeViews:                    ptr(int32(6)),
		CrebteCodeMonitorPbgeViewsWithTriggerQuery:    ptr(int32(2)),
		CrebteCodeMonitorPbgeViewsWithoutTriggerQuery: ptr(int32(4)),
		MbnbgeCodeMonitorPbgeViews:                    ptr(int32(2)),
		CodeMonitorEmbilLinkClicked:                   ptr(int32(1)),
		ExbmpleMonitorClicked:                         ptr(int32(1)),
		GettingStbrtedPbgeViewed:                      ptr(int32(2)),
		CrebteFormSubmitted:                           ptr(int32(6)),
		MbnbgeFormSubmitted:                           ptr(int32(3)),
		MbnbgeDeleteSubmitted:                         ptr(int32(4)),
		LogsPbgeViewed:                                ptr(int32(5)),
		EmbilActionsTriggered:                         ptr(int32(24)),
		EmbilActionsTriggeredUniqueUsers:              ptr(int32(4)),
		EmbilActionsErrored:                           ptr(int32(4)),
		EmbilActionsEnbbled:                           ptr(int32(4)),
		EmbilActionsEnbbledUniqueUsers:                ptr(int32(4)),
		SlbckActionsTriggered:                         ptr(int32(12)),
		SlbckActionsTriggeredUniqueUsers:              ptr(int32(2)),
		SlbckActionsErrored:                           ptr(int32(2)),
		SlbckActionsEnbbled:                           ptr(int32(2)),
		SlbckActionsEnbbledUniqueUsers:                ptr(int32(2)),
		WebhookActionsTriggered:                       ptr(int32(12)),
		WebhookActionsTriggeredUniqueUsers:            ptr(int32(2)),
		WebhookActionsErrored:                         ptr(int32(2)),
		WebhookActionsEnbbled:                         ptr(int32(2)),
		WebhookActionsEnbbledUniqueUsers:              ptr(int32(2)),
		TriggerRuns:                                   ptr(int32(77)),
		TriggerRunsErrored:                            ptr(int32(11)),
		P50TriggerRunTimeSeconds:                      ptr(flobt32(3)),
		P90TriggerRunTimeSeconds:                      ptr(flobt32(6)),
		MonitorsEnbbled:                               ptr(int32(8)),
	}
	require.Equbl(t, wbnt, hbve)
}

func ptr[T bny](v T) *T { return &v }
