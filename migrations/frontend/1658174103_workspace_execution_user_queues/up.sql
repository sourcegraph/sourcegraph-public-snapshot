CREATE TABLE IF NOT EXISTS batch_spec_workspace_execution_last_dequeues (
    user_id integer PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY DEFERRED,
    latest_dequeue timestamp with time zone
);

CREATE OR REPLACE FUNCTION batch_spec_workspace_execution_last_dequeues_upsert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
    INSERT INTO
        batch_spec_workspace_execution_last_dequeues
    SELECT
        user_id,
        MAX(started_at) as latest_dequeue
    FROM
        newtab
    GROUP BY
        user_id
    ON CONFLICT (user_id) DO UPDATE SET
        latest_dequeue = GREATEST(batch_spec_workspace_execution_last_dequeues.latest_dequeue, EXCLUDED.latest_dequeue);

    RETURN NULL;
END $$;

DROP TRIGGER IF EXISTS batch_spec_workspace_execution_last_dequeues_insert ON batch_spec_workspace_execution_jobs;
CREATE TRIGGER batch_spec_workspace_execution_last_dequeues_insert AFTER INSERT ON batch_spec_workspace_execution_jobs REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION batch_spec_workspace_execution_last_dequeues_upsert();

DROP TRIGGER IF EXISTS batch_spec_workspace_execution_last_dequeues_update ON batch_spec_workspace_execution_jobs;
CREATE TRIGGER batch_spec_workspace_execution_last_dequeues_update AFTER UPDATE ON batch_spec_workspace_execution_jobs REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION batch_spec_workspace_execution_last_dequeues_upsert();

DROP VIEW IF EXISTS batch_spec_workspace_execution_jobs_with_rank;
DROP VIEW IF EXISTS batch_spec_workspace_execution_queue;

CREATE VIEW batch_spec_workspace_execution_queue AS
WITH queue_candidates AS (
    SELECT
        exec.id,
        RANK() OVER (
            PARTITION BY queue.user_id
            -- Make sure the jobs are still fulfilled in timely order, and that the ordering is stable.
            ORDER BY exec.created_at ASC, exec.id ASC
        ) AS place_in_user_queue
    FROM batch_spec_workspace_execution_jobs exec
    JOIN batch_spec_workspace_execution_last_dequeues queue ON queue.user_id = exec.user_id
    WHERE
    	-- Only queued records should get a rank.
        exec.state = 'queued'
    ORDER BY
        -- Round-robin let users dequeue jobs.
        place_in_user_queue,
        -- And ensure the user who dequeued the longest ago is next.
        queue.latest_dequeue ASC NULLS FIRST
)
SELECT
    queue_candidates.id, ROW_NUMBER() OVER () AS place_in_global_queue, queue_candidates.place_in_user_queue
FROM queue_candidates;

CREATE VIEW batch_spec_workspace_execution_jobs_with_rank AS (
    SELECT
        j.*,
        q.place_in_global_queue,
        q.place_in_user_queue
    FROM
        batch_spec_workspace_execution_jobs j
    LEFT JOIN batch_spec_workspace_execution_queue q ON j.id = q.id
);

INSERT INTO batch_spec_workspace_execution_last_dequeues
SELECT
    exec.user_id as user_id,
    MAX(exec.started_at) AS latest_dequeue
FROM batch_spec_workspace_execution_jobs exec
GROUP BY exec.user_id
ON CONFLICT (user_id) DO UPDATE
    SET latest_dequeue = GREATEST(batch_spec_workspace_execution_last_dequeues.latest_dequeue, EXCLUDED.latest_dequeue);
