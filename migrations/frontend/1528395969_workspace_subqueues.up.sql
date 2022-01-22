-- +++
-- parent: 1528395968
-- +++

BEGIN;

CREATE TABLE IF NOT EXISTS batch_spec_workspace_execution_maximum_concurrencies (
    user_id INTEGER NOT NULL PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE DEFERRABLE,
    concurrency INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS batch_spec_workspace_execution_hourly_quota (
    user_id INTEGER NOT NULL PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE DEFERRABLE,
    quota INTEGER NOT NULL
);

CREATE VIEW IF NOT EXISTS execution_jobs AS (
    SELECT
        j.id,
        j.started_at,
        j.process_after,
        j.finished_at,
        j.created_at,
        j.state,
        b.user_id,
        ROW_NUMBER() OVER (PARTITION BY b.user_id ORDER BY j.started_at, j.process_after) AS rank
    FROM batch_spec_workspace_execution_jobs j
    JOIN batch_spec_workspaces w ON w.id = j.batch_spec_workspace_id
    JOIN batch_specs b ON b.id = w.batch_spec_id;
);

-- TODO: We need to make sure that execution jobs are never deleted to not lose
-- track of dequeued jobs.
CREATE VIEW IF NOT EXISTS batch_spec_workspace_execution_jobs_subqueues AS (
    -- First generate the tenant queues by getting all jobs grouped by the tenant (user) id.
    WITH tenant_queues AS (
    SELECT
        j.user_id,
        COUNT(1) FILTER (WHERE state = 'queued') as queue_length,
        COUNT(1) FILTER (WHERE state = 'processing') as current_concurrency,
        -- The number of jobs that successfully dequeued in the past hour.
        -- This does not include jobs that restarted due to error.
        COUNT(1) FILTER (WHERE started_at IS NOT NULL AND started_at > NOW() - interval '1 hour') as hourly_dequeue_rate,
        -- The last time this tenant was able to dequeue a job.
        MAX(j.started_at) as latest_dequeue
    FROM execution_jobs j
    GROUP BY j.user_id
    ),
    tenant_info AS (
        SELECT
            q.user_id,
            q.queue_length,
            COALESCE(
                (select concurrency from batch_spec_workspace_execution_maximum_concurrencies mc where mc.user_id = q.user_id),
                -- The default if nothing is configured.
                4
            ) as max_concurrency,
            COALESCE(
            (select quota from batch_spec_workspace_execution_hourly_quota hq where hq.user_id = q.user_id),
                -- The default if nothing is configured.
                500
            ) as hourly_quota,
            q.current_concurrency,
            q.hourly_dequeue_rate,
            q.latest_dequeue,
            j.id as job_id,
            ROW_NUMBER() OVER (
                PARTITION BY q.user_id
                ORDER BY
                    -- Still dequeue in order of creation (todo: readd order by crated_at, process_after)
                    j.created_at ASC, process_after ASC
            ) AS rank
        FROM execution_jobs j
        JOIN tenant_queues q ON q.user_id = j.user_id
        WHERE
            -- Only queued and errored jobs are dequeuable.
            (j.state = 'queued' OR j.state = 'errored')

            AND

            -- And only if their process_after date has passed yet.
            (j.process_after IS NULL OR j.process_after <= NOW())
    ),
    candidates AS (
        SELECT job_id, rank, latest_dequeue from tenant_info
        WHERE
            -- Limit concurrency: Maximum sub-queue concurrency - running jobs - position in queue needs to be >= 0.
            max_concurrency - current_concurrency - rank >= 0
            AND
            -- Limit also on the maximum number of jobs permitted per hour.
            hourly_dequeue_rate < hourly_quota
    )
    SELECT
        batch_spec_workspace_execution_jobs.*,
        rank,
        latest_dequeue
    FROM
        batch_spec_workspace_execution_jobs
    -- Must exist in the list of candidates.
    JOIN candidates ON candidates.job_id = batch_spec_workspace_execution_jobs.id
);

COMMIT;
