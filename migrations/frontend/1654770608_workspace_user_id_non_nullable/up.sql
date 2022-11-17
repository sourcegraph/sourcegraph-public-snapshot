UPDATE batch_spec_workspace_execution_jobs exec SET user_id = (
    SELECT spec.user_id
    FROM batch_spec_workspaces AS workspace
    JOIN batch_specs spec ON spec.id = workspace.batch_spec_id
    WHERE workspace.id = exec.batch_spec_workspace_id
);

ALTER TABLE batch_spec_workspace_execution_jobs ALTER COLUMN user_id SET NOT NULL;
