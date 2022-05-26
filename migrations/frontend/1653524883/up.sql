CREATE VIEW batch_spec_workspace_execution_queue AS
WITH tenant_queues AS (
    SELECT
        spec.user_id,
        COUNT(1) FILTER (WHERE exec.state = 'queued') as queue_length,
        COUNT(1) FILTER (WHERE exec.state = 'processing') as current_concurrency,
        MAX(exec.started_at) as latest_dequeue
    FROM batch_spec_workspace_execution_jobs AS exec
    JOIN batch_spec_workspaces AS workspace ON workspace.id = exec.batch_spec_workspace_id
    JOIN batch_specs spec ON spec.id = workspace.batch_spec_id
    GROUP BY spec.user_id
)
SELECT
    exec.*,
    queue.queue_length,
    queue.current_concurrency,
    queue.latest_dequeue
FROM batch_spec_workspace_execution_jobs exec
JOIN batch_spec_workspaces workspace ON workspace.id = exec.batch_spec_workspace_id
JOIN batch_specs spec ON spec.id = workspace.batch_spec_id
JOIN tenant_queues queue ON queue.user_id = spec.user_id
WHERE
	queue.current_concurrency < 4 AND exec.state = 'queued'
ORDER BY
	-- Round-robin let tenants dequeue jobs
	ROW_NUMBER() OVER (
		PARTITION BY queue.user_id
		ORDER BY
			queue.latest_dequeue ASC NULLS FIRST,
			exec.id
	);
