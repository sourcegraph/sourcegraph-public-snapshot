WITH event_log_stats AS (
    SELECT
        NULLIF(COUNT(*) FILTER (WHERE name IN ('ViewCodeMonitoringPage', 'CodeMonitoringPageViewed')), 0) :: INT AS code_monitoring_page_views,
        NULLIF(COUNT(*) FILTER (WHERE name IN ('ViewCreateCodeMonitorPage', 'CreateCodeMonitorPageViewed')), 0) :: INT AS create_code_monitor_page_views,
        NULLIF(COUNT(*) FILTER (WHERE name IN ('ViewCreateCodeMonitorPage', 'CreateCodeMonitorPageViewed') AND (argument->>'hasTriggerQuery')::bool), 0) :: INT AS create_code_monitor_page_views_with_trigger_query,
        NULLIF(COUNT(*) FILTER (WHERE name IN ('ViewCreateCodeMonitorPage', 'CreateCodeMonitorPageViewed') AND NOT (argument->>'hasTriggerQuery')::bool), 0) :: INT AS create_code_monitor_page_views_without_trigger_query,
        NULLIF(COUNT(*) FILTER (WHERE name IN ('ViewManageCodeMonitorPage', 'ManageCodeMonitorPageViewed')), 0) :: INT AS manage_code_monitor_page_views,
        NULLIF(COUNT(*) FILTER (WHERE name = 'CodeMonitorEmailLinkClicked'), 0) :: INT AS code_monitor_email_link_clicks,
        NULLIF(COUNT(*) FILTER (WHERE name = 'CodeMonitoringExampleMonitorClicked'), 0) :: INT AS code_monitor_example_monitor_clicked,
        NULLIF(COUNT(*) FILTER (WHERE name = 'CodeMonitoringGettingStartedPageViewed'), 0) :: INT AS code_monitor_getting_started_page_viewed,
        NULLIF(COUNT(*) FILTER (WHERE name = 'CreateCodeMonitorFormSubmitted'), 0) :: INT AS code_monitor_create_form_submitted,
        NULLIF(COUNT(*) FILTER (WHERE name = 'ManageCodeMonitorFormSubmitted'), 0) :: INT AS code_monitor_manage_form_submitted,
        NULLIF(COUNT(*) FILTER (WHERE name = 'ManageCodeMonitorDeleteSubmitted'), 0) :: INT AS code_monitor_manage_delete_submitted,
        NULLIF(COUNT(*) FILTER (WHERE name = 'CodeMonitoringLogsPageViewed'), 0) :: INT AS code_monitor_logs_page_viewed
    FROM event_logs
    WHERE
        name IN (
            -- The events that share a line are events that changed names and are aliases of each other
            'ViewCodeMonitoringPage', 'CodeMonitoringPageViewed',
            'ViewCreateCodeMonitorPage', 'CreateCodeMonitorPageViewed',
            'ViewManageCodeMonitorPage', 'ManageCodeMonitorPageViewed',
            'CodeMonitorEmailLinkClicked',
            'CodeMonitoringExampleMonitorClicked',
            'CodeMonitoringGettingStartedPageViewed',
            'CreateCodeMonitorFormSubmitted',
            'ManageCodeMonitorFormSubmitted',
            'ManageCodeMonitorDeleteSubmitted',
            'CodeMonitoringLogsPageViewed'
        )
), email_actions AS (
	SELECT
        NULLIF(COUNT(*), 0) :: INT AS email_actions_enabled,
        NULLIF(COUNT(DISTINCT users.id), 0) :: INT as email_actions_enabled_unique_users
	FROM cm_emails
    LEFT JOIN cm_monitors ON cm_emails.monitor = cm_monitors.id
    LEFT JOIN users ON cm_monitors.namespace_user_id = users.id
    WHERE cm_emails.enabled AND cm_monitors.enabled
), slack_actions AS (
	SELECT
        NULLIF(COUNT(*), 0) :: INT AS slack_actions_enabled,
        NULLIF(COUNT(DISTINCT users.id), 0) :: INT as slack_actions_enabled_unique_users
	FROM cm_slack_webhooks
    LEFT JOIN cm_monitors ON cm_slack_webhooks.monitor = cm_monitors.id
    LEFT JOIN users ON cm_monitors.namespace_user_id = users.id
    WHERE cm_slack_webhooks.enabled AND cm_monitors.enabled
), webhook_actions AS (
	SELECT
        NULLIF(COUNT(*), 0) :: INT AS webhook_actions_enabled,
        NULLIF(COUNT(DISTINCT users.id), 0) :: INT as webhook_actions_enabled_unique_users
	FROM cm_webhooks
    LEFT JOIN cm_monitors ON cm_webhooks.monitor = cm_monitors.id
    LEFT JOIN users ON cm_monitors.namespace_user_id = users.id
    WHERE cm_webhooks.enabled AND cm_monitors.enabled
), action_jobs AS (
    SELECT
        NULLIF(COUNT(*) FILTER (WHERE email IS NOT NULL), 0) :: INT AS email_actions_triggered,
        NULLIF(COUNT(DISTINCT users.id) FILTER (WHERE email IS NOT NULL), 0) :: INT AS email_actions_triggered_unique_users,
        NULLIF(COUNT(*) FILTER (WHERE email IS NOT NULL AND state = 'failed'), 0) :: INT AS email_actions_errored,
        NULLIF(COUNT(*) FILTER (WHERE slack_webhook IS NOT NULL), 0) :: INT AS slack_actions_triggered,
        NULLIF(COUNT(DISTINCT users.id) FILTER (WHERE slack_webhook IS NOT NULL), 0) :: INT AS slack_actions_triggered_unique_users,
        NULLIF(COUNT(*) FILTER (WHERE slack_webhook IS NOT NULL AND state = 'failed'), 0) :: INT AS slack_actions_errored,
        NULLIF(COUNT(*) FILTER (WHERE webhook IS NOT NULL), 0) :: INT AS webhook_actions_triggered,
        NULLIF(COUNT(DISTINCT users.id) FILTER (WHERE webhook IS NOT NULL), 0) :: INT AS webhook_actions_triggered_unique_users,
        NULLIF(COUNT(*) FILTER (WHERE webhook IS NOT NULL AND state = 'failed'), 0) :: INT AS webhook_actions_errored
    FROM cm_action_jobs
    LEFT JOIN cm_emails ON cm_emails.id = cm_action_jobs.email
    LEFT JOIN cm_slack_webhooks ON cm_slack_webhooks.id = cm_action_jobs.slack_webhook
    LEFT JOIN cm_webhooks ON cm_webhooks.id = cm_action_jobs.webhook
    LEFT JOIN cm_monitors ON cm_monitors.id = COALESCE(cm_emails.monitor, cm_slack_webhooks.monitor, cm_webhooks.monitor)
    LEFT JOIN users ON cm_monitors.namespace_user_id = users.id
    WHERE cm_action_jobs.started_at > now() - '1 week'::interval
), trigger_jobs AS (
    SELECT
        NULLIF(COUNT(*), 0) :: INT AS trigger_runs,
        NULLIF(COUNT(*) FILTER (WHERE state = 'failed'), 0) :: INT AS trigger_runs_errored,
        PERCENTILE_CONT(0.5)
            WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (finished_at - started_at)))
            FILTER (WHERE finished_at IS NOT NULL and started_at IS NOT NULL)
            AS p50_trigger_run_time,
        PERCENTILE_CONT(0.9)
            WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (finished_at - started_at)))
            FILTER (WHERE finished_at IS NOT NULL and started_at IS NOT NULL)
            AS p90_trigger_run_time
    FROM cm_trigger_jobs
    LEFT JOIN cm_queries ON cm_queries.id = cm_trigger_jobs.query
    LEFT JOIN cm_monitors ON cm_monitors.id = cm_queries.monitor
    LEFT JOIN users ON cm_monitors.namespace_user_id = users.id
    WHERE cm_trigger_jobs.started_at > now() - '1 week'::interval
), monitored_repos AS (
    SELECT
        COUNT(DISTINCT repo_id) as repos_monitored
    FROM cm_last_searched
)
SELECT
    event_log_stats.code_monitoring_page_views,
    event_log_stats.create_code_monitor_page_views,
    event_log_stats.create_code_monitor_page_views_with_trigger_query,
    event_log_stats.create_code_monitor_page_views_without_trigger_query,
    event_log_stats.manage_code_monitor_page_views,
    event_log_stats.code_monitor_email_link_clicks,
    event_log_stats.code_monitor_example_monitor_clicked,
    event_log_stats.code_monitor_getting_started_page_viewed,
    event_log_stats.code_monitor_create_form_submitted,
    event_log_stats.code_monitor_manage_form_submitted,
    event_log_stats.code_monitor_manage_delete_submitted,
    event_log_stats.code_monitor_logs_page_viewed,
    email_actions.email_actions_enabled,
    email_actions.email_actions_enabled_unique_users,
    slack_actions.slack_actions_enabled,
    slack_actions.slack_actions_enabled_unique_users,
    webhook_actions.webhook_actions_enabled,
    webhook_actions.webhook_actions_enabled_unique_users,
    action_jobs.email_actions_triggered,
    action_jobs.email_actions_triggered_unique_users,
    action_jobs.email_actions_errored,
    action_jobs.slack_actions_triggered,
    action_jobs.slack_actions_triggered_unique_users,
    action_jobs.slack_actions_errored,
    action_jobs.webhook_actions_triggered,
    action_jobs.webhook_actions_triggered_unique_users,
    action_jobs.webhook_actions_errored,
    trigger_jobs.trigger_runs,
    trigger_jobs.trigger_runs_errored,
    trigger_jobs.p50_trigger_run_time,
    trigger_jobs.p90_trigger_run_time,
    -- monitors_enabled
    COALESCE(slack_actions.slack_actions_enabled, 0) +
    COALESCE(email_actions.email_actions_enabled, 0) +
    COALESCE(webhook_actions.webhook_actions_enabled, 0) AS monitors_enabled,
    -- monitors_enabled_unique_users
    COALESCE(slack_actions.slack_actions_enabled_unique_users, 0) +
    COALESCE(email_actions.email_actions_enabled_unique_users, 0) +
    COALESCE(webhook_actions.webhook_actions_enabled_unique_users, 0) AS monitors_enabled_unique_users
FROM
    event_log_stats,
    email_actions,
    slack_actions,
    webhook_actions,
    action_jobs,
    trigger_jobs
;
