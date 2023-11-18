package usagestats

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCodeMonitoringUsageStatistics(t *testing.T) {
	ctx := context.Background()

	now := time.Date(2021, 1, 28, 0, 0, 0, 0, time.UTC)
	mockTimeNow(now)
	defer func() {
		timeNow = time.Now
	}()

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	_, err := db.ExecContext(ctx, `
		INSERT INTO event_logs
			(name, argument, url, user_id, anonymous_user_id, source, version, timestamp)
		VALUES
			('ViewCodeMonitoringPage', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CodeMonitoringPageViewed', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('ViewCreateCodeMonitorPage', '{"hasTriggerQuery": false}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('ViewCreateCodeMonitorPage', '{"hasTriggerQuery": false}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('ViewCreateCodeMonitorPage', '{"hasTriggerQuery": true}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CreateCodeMonitorPageViewed', '{"hasTriggerQuery": false}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CreateCodeMonitorPageViewed', '{"hasTriggerQuery": false}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CreateCodeMonitorPageViewed', '{"hasTriggerQuery": true}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('ViewManageCodeMonitorPage', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('ManageCodeMonitorPageViewed', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CodeMonitorEmailLinkClicked', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			-- x1
			('CodeMonitoringExampleMonitorClicked', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			-- x2
			('CodeMonitoringGettingStartedPageViewed', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CodeMonitoringGettingStartedPageViewed', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			-- x3
			('ManageCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('ManageCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('ManageCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			-- x4
			('ManageCodeMonitorDeleteSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('ManageCodeMonitorDeleteSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('ManageCodeMonitorDeleteSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('ManageCodeMonitorDeleteSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			-- x5
			('CodeMonitoringLogsPageViewed', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CodeMonitoringLogsPageViewed', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CodeMonitoringLogsPageViewed', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CodeMonitoringLogsPageViewed', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CodeMonitoringLogsPageViewed', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			-- x6
			('CreateCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CreateCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CreateCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CreateCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CreateCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			('CreateCodeMonitorFormSubmitted', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day')
	`, now)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, username)
		VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cm_monitors (id, enabled, created_by, changed_by, namespace_user_id, description)
		VALUES
			-- User 1 has 2 monitors of each action type, all enabled
			(1, true,  1, 1, 1, ''),
			(2, true,  1, 1, 1, ''),
			(3, true,  1, 1, 1, ''),
			(4, true,  1, 1, 1, ''),
			(5, true,  1, 1, 1, ''),
			(6, true,  1, 1, 1, ''),
			-- User 2 has 2 monitors of each action type, half disabled
			(7,  true,  2, 2, 2, ''),
			(8,  false, 2, 2, 2, ''),
			(9,  true,  2, 2, 2, ''),
			(10, false, 2, 2, 2, ''),
			(11, true,  2, 2, 2, ''),
			(12, false, 2, 2, 2, ''),
			-- User 3 has 1 monitor, enabled
			(13, true, 3, 3, 3, ''),
			-- User 4 has 1 monitor, disabled
			(14, true, 4, 4, 4, '')
			-- User 5 has no monitors
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cm_queries (monitor, query, created_by, changed_by)
		SELECT
			s as monitor,
			'',
			cm_monitors.created_by,
			cm_monitors.changed_by
		FROM generate_series(1,14) s
		JOIN cm_monitors ON cm_monitors.id = s
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cm_emails
			(monitor, enabled, priority, header, created_by, changed_by)
		SELECT
			cm_monitors.id,
			CASE WHEN cm_monitors.id = 2 THEN false ELSE true END,
			'NORMAL',
			'',
			cm_monitors.created_by,
			cm_monitors.changed_by
		FROM cm_monitors
		WHERE cm_monitors.id IN (1, 2, 7, 8, 13, 14)
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cm_slack_webhooks
			(monitor, enabled, url, created_by, changed_by)
		SELECT
			cm_monitors.id,
			CASE WHEN cm_monitors.id = 4 THEN false ELSE true END,
			'',
			cm_monitors.created_by,
			cm_monitors.changed_by
		FROM cm_monitors
		WHERE cm_monitors.id IN (3, 4, 9, 10);
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cm_webhooks
			(monitor, enabled, url, created_by, changed_by)
		SELECT
			cm_monitors.id,
			CASE WHEN cm_monitors.id = 6 THEN false ELSE true END,
			'',
			cm_monitors.created_by,
			cm_monitors.changed_by
		FROM cm_monitors
		WHERE cm_monitors.id IN (5, 6, 11, 12)
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cm_trigger_jobs (query, state, finished_at, started_at)
		SELECT
			cm_queries.id,
			CASE WHEN s < 6 THEN 'completed' ELSE 'failed' END,
			now() - s * '1 day'::interval,
			now() - s * '1 day'::interval - s * '1 second'::interval
		FROM cm_queries
		JOIN cm_monitors ON cm_queries.monitor = cm_monitors.id
		CROSS JOIN generate_series(0, 10) s
		WHERE cm_monitors.enabled
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cm_action_jobs
			(email, webhook, slack_webhook, state, finished_at, started_at)
		SELECT
			cm_emails.id,
			NULL::bigint,
			NULL::bigint,
			CASE WHEN s < 6 THEN 'completed' ELSE 'failed' END,
			now() - s * '1 day'::interval,
			now() - s * '1 day'::interval - s * '1 second'::interval
		FROM cm_emails
		CROSS JOIN generate_series(1, 10) s
		JOIN cm_monitors ON cm_emails.monitor = cm_monitors.id
		WHERE cm_emails.enabled
			AND cm_monitors.enabled
		UNION ALL
		SELECT
			NULL::bigint,
			cm_webhooks.id,
			NULL::bigint,
			CASE WHEN s < 6 THEN 'completed' ELSE 'failed' END,
			now() - s * '1 day'::interval,
			now() - s * '1 day'::interval - s * '1 second'::interval
		FROM cm_webhooks
		CROSS JOIN generate_series(1, 10) s
		JOIN cm_monitors ON cm_webhooks.monitor = cm_monitors.id
		WHERE cm_webhooks.enabled
			AND cm_monitors.enabled
		UNION ALL
		SELECT
			NULL::bigint,
			NULL::bigint,
			cm_slack_webhooks.id,
			CASE WHEN s < 6 THEN 'completed' ELSE 'failed' END,
			now() - s * '1 day'::interval,
			now() - s * '1 day'::interval - s * '1 second'::interval
		FROM cm_slack_webhooks
		CROSS JOIN generate_series(1, 10) s
		JOIN cm_monitors ON cm_slack_webhooks.monitor = cm_monitors.id
		WHERE cm_slack_webhooks.enabled
			AND cm_monitors.enabled
	`)
	require.NoError(t, err)

	have, err := GetCodeMonitoringUsageStatistics(ctx, db)
	require.NoError(t, err)

	want := &types.CodeMonitoringUsageStatistics{
		CodeMonitoringPageViews:                       ptr(int32(2)),
		CreateCodeMonitorPageViews:                    ptr(int32(6)),
		CreateCodeMonitorPageViewsWithTriggerQuery:    ptr(int32(2)),
		CreateCodeMonitorPageViewsWithoutTriggerQuery: ptr(int32(4)),
		ManageCodeMonitorPageViews:                    ptr(int32(2)),
		CodeMonitorEmailLinkClicked:                   ptr(int32(1)),
		ExampleMonitorClicked:                         ptr(int32(1)),
		GettingStartedPageViewed:                      ptr(int32(2)),
		CreateFormSubmitted:                           ptr(int32(6)),
		ManageFormSubmitted:                           ptr(int32(3)),
		ManageDeleteSubmitted:                         ptr(int32(4)),
		LogsPageViewed:                                ptr(int32(5)),
		EmailActionsTriggered:                         ptr(int32(24)),
		EmailActionsTriggeredUniqueUsers:              ptr(int32(4)),
		EmailActionsErrored:                           ptr(int32(4)),
		EmailActionsEnabled:                           ptr(int32(4)),
		EmailActionsEnabledUniqueUsers:                ptr(int32(4)),
		SlackActionsTriggered:                         ptr(int32(12)),
		SlackActionsTriggeredUniqueUsers:              ptr(int32(2)),
		SlackActionsErrored:                           ptr(int32(2)),
		SlackActionsEnabled:                           ptr(int32(2)),
		SlackActionsEnabledUniqueUsers:                ptr(int32(2)),
		WebhookActionsTriggered:                       ptr(int32(12)),
		WebhookActionsTriggeredUniqueUsers:            ptr(int32(2)),
		WebhookActionsErrored:                         ptr(int32(2)),
		WebhookActionsEnabled:                         ptr(int32(2)),
		WebhookActionsEnabledUniqueUsers:              ptr(int32(2)),
		TriggerRuns:                                   ptr(int32(77)),
		TriggerRunsErrored:                            ptr(int32(11)),
		P50TriggerRunTimeSeconds:                      ptr(float32(3)),
		P90TriggerRunTimeSeconds:                      ptr(float32(6)),
		MonitorsEnabled:                               ptr(int32(8)),
		MonitorsEnabledUniqueUsers:                    ptr(int32(8)),
	}
	require.Equal(t, want, have)
}

func ptr[T any](v T) *T { return &v }
