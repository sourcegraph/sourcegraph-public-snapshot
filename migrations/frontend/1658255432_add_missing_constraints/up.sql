DELETE FROM batch_spec_resolution_jobs WHERE batch_spec_id IS NULL OR state IS NULL;

ALTER TABLE batch_spec_resolution_jobs
    ALTER COLUMN batch_spec_id SET NOT NULL,
    ALTER COLUMN state SET NOT NULL;

DELETE FROM batch_spec_workspace_execution_jobs WHERE batch_spec_workspace_id IS NULL OR state IS NULL;

ALTER TABLE batch_spec_workspace_execution_jobs
    ALTER COLUMN batch_spec_workspace_id SET NOT NULL,
    ALTER COLUMN state SET NOT NULL;

DELETE FROM batch_spec_workspaces WHERE batch_spec_id IS NULL OR changeset_spec_ids IS NULL OR repo_id IS NULL;

ALTER TABLE batch_spec_workspaces
    ALTER COLUMN batch_spec_id SET NOT NULL,
    ALTER COLUMN changeset_spec_ids SET NOT NULL,
    ALTER COLUMN repo_id SET NOT NULL;

DELETE FROM changeset_jobs WHERE state IS NULL;
ALTER TABLE changeset_jobs
    ALTER COLUMN state SET NOT NULL;
