-- +++
-- parent: 1528395968
-- +++

BEGIN;

CREATE TABLE IF NOT EXISTS batch_spec_workspace_execution_maximum_concurrencies (
    -- TODO: Foreign key constraint.
    user_id integer NOT NULL,
    concurrency integer NOT NULL
);

CREATE VIEW batch_spec_workspace_execution_jobs_subqueues AS (
    WITH tenant_queues AS (
    SELECT
        j.user_id,
        COUNT(1) FILTER (WHERE state = 'queued') as queue_length,
        COUNT(1) FILTER (WHERE state = 'processing') as current_concurrency,
        COUNT(1) FILTER (WHERE started_at IS NOT NULL AND started_at > NOW() - interval '1 hour') as hourly_dequeue_rate,
        MAX(j.started_at) as latest_dequeue
    FROM execution_jobs j
    GROUP BY j.user_id
    ), tenant_info AS (
    SELECT
        q.user_id,
        q.queue_length,
        COALESCE(
            (select concurrency from batch_spec_workspace_execution_maximum_concurrencies where batch_spec_workspace_execution_maximum_concurrencies.user_id = q.user_id),
            -- The default if nothing is configured.
            4
        ) as max_concurrency,
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
        (j.state = 'queued' OR j.state = 'errored')

        AND

        (j.process_after IS NULL OR j.process_after <= NOW())
    )
    , candidates AS (
        SELECT job_id, rank, latest_dequeue from tenant_info
        WHERE
            -- Limit concurrency: Maximum sub-queue concurrency - running jobs - position in queue needs to be >= 0.
            max_concurrency - current_concurrency - rank >= 0
            AND
            -- Limit also on the maximum number of jobs permitted per hour. TODO: Make configurable.
            hourly_dequeue_rate < 500
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
