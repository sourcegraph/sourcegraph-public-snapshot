WITH event_log_stats AS (
    SELECT
        NULLIF(COUNT(*) FILTER (WHERE name IN ('ViewCodeMonitoringPage', 'CodeMonitoringPageViewed')), 0) :: INT AS code_monitoring_page_views,
    FROM event_logs
    WHERE
        name IN (
            -- The events that share a line are events that changed names and are aliases of each other
            'CodeMonitoringLogsPageViewed'
        )
)
SELECT
    event_log_stats.code_monitoring_page_views,
    event_log_stats.create_code_monitor_page_views,
    trigger_jobs.p90_trigger_run_time
FROM
    event_log_stats,
    email_actions,
    slack_actions,
    webhook_actions,
    action_jobs,
    trigger_jobs
;
