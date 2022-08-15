ALTER TABLE batch_spec_workspace_execution_jobs ADD COLUMN IF NOT EXISTS access_token_id integer REFERENCES access_tokens(id);

DROP VIEW IF EXISTS batch_spec_workspace_execution_jobs_with_rank;
CREATE VIEW batch_spec_workspace_execution_jobs_with_rank AS (
    SELECT
        j.*,
        q.place_in_global_queue,
        q.place_in_user_queue
    FROM
        batch_spec_workspace_execution_jobs j
    LEFT JOIN batch_spec_workspace_execution_queue q ON j.id = q.id
);
