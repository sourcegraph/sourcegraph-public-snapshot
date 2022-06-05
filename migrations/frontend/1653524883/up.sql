DROP VIEW IF EXISTS batch_spec_workspace_execution_queue;

CREATE VIEW batch_spec_workspace_execution_queue AS
WITH tenant_queues AS (
    SELECT
        spec.user_id,
        MAX(exec.started_at) AS latest_dequeue
    FROM batch_spec_workspace_execution_jobs AS exec
    JOIN batch_spec_workspaces AS workspace ON workspace.id = exec.batch_spec_workspace_id
    JOIN batch_specs spec ON spec.id = workspace.batch_spec_id
    GROUP BY spec.user_id
),
-- We are creating this materialized CTE because PostgreSQL doesn't allow `FOR UPDATE` with window functions.
-- Materializing it makes sure that the view query is not inlined into the FOR UPDATE select the Dequeue method
-- performs.
materialized_queue_candidates AS MATERIALIZED (
    SELECT
        exec.*,
        RANK() OVER (
            PARTITION BY queue.user_id
            -- Make sure the jobs are still fulfilled in timely order, and that the ordering is stable.
            ORDER BY exec.created_at ASC, exec.id ASC
        ) AS tenant_queue_rank
    FROM batch_spec_workspace_execution_jobs exec
    -- Join workspaces, because we need to map exec->tenant_queue via exec_jobs->workspaces->batch_specs->tenant_queues, phew.
    -- Optimization: Store the tenant on the job record directly, although it's a denormalization.
    JOIN batch_spec_workspaces workspace ON workspace.id = exec.batch_spec_workspace_id
    JOIN batch_specs spec ON spec.id = workspace.batch_spec_id
    JOIN tenant_queues queue ON queue.user_id = spec.user_id
    WHERE
    	-- Only queued records should get a rank.
        exec.state = 'queued'
    ORDER BY
        -- Round-robin let tenants dequeue jobs.
        tenant_queue_rank,
        -- And ensure the user who dequeued the longest ago is next.
        queue.latest_dequeue ASC NULLS FIRST
)
SELECT
    ROW_NUMBER() OVER () AS place_in_queue, materialized_queue_candidates.*
FROM materialized_queue_candidates;
