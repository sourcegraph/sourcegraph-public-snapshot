ALTER TABLE batch_spec_resolution_jobs
    ALTER COLUMN batch_spec_id DROP NOT NULL,
    ALTER COLUMN state DROP NOT NULL;

ALTER TABLE batch_spec_workspace_execution_jobs
    ALTER COLUMN batch_spec_workspace_id DROP NOT NULL,
    ALTER COLUMN state DROP NOT NULL;

ALTER TABLE batch_spec_workspaces
    ALTER COLUMN batch_spec_id DROP NOT NULL,
    ALTER COLUMN changeset_spec_ids DROP NOT NULL,
    ALTER COLUMN repo_id DROP NOT NULL;

ALTER TABLE changeset_jobs
    ALTER COLUMN state DROP NOT NULL;
