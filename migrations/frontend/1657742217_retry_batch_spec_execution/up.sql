CREATE OR REPLACE FUNCTION func_retry_batch_spec_execution(input_batch_spec_id bigint) RETURNS VOID AS
$$
DECLARE
    workspace_ids                bigint[];
    workspace_changeset_spec_ids bigint[];
BEGIN
    -- Prevent retrying already applied batch changes
    IF exists(SELECT 1
              FROM batch_changes
              WHERE batch_spec_id = input_batch_spec_id
                AND last_applied_at IS NOT NULL
        ) THEN
        RAISE EXCEPTION 'batch spec already applied';
    END IF;
    -- Get workspace ids
    SELECT array_agg(DISTINCT batch_spec_workspaces.id),
           array_agg(DISTINCT spec_ids)
    INTO workspace_ids, workspace_changeset_spec_ids
    FROM batch_spec_workspaces
             INNER JOIN repo ON repo.id = batch_spec_workspaces.repo_id
             INNER JOIN batch_spec_workspace_execution_jobs
                        on batch_spec_workspaces.id =
                           batch_spec_workspace_execution_jobs.batch_spec_workspace_id
        -- Need to put the changeset_spec_ids into an array, but need to unwrap the array inorder to aggregate
             LEFT JOIN LATERAL unnest(array(SELECT jsonb_object_keys(batch_spec_workspaces.changeset_spec_ids)::bigint)) WITH ORDINALITY AS spec_ids
                       ON TRUE
    WHERE repo.deleted_at IS NULL
      AND batch_spec_workspaces.batch_spec_id = input_batch_spec_id
      AND batch_spec_workspace_execution_jobs.state != 'completed';
    -- Cleanup any uncompleted execution jobs
    DELETE
    FROM batch_spec_workspace_execution_jobs
    WHERE batch_spec_workspace_execution_jobs.batch_spec_workspace_id = ANY (workspace_ids);
    -- Cleanup any changeset specs that were created
    IF array_length(workspace_changeset_spec_ids, 1) > 0 THEN
        DELETE
        FROM changeset_specs
        WHERE changeset_specs.id = ANY (workspace_changeset_spec_ids);
    END IF;
    -- Create new jobs
    INSERT INTO batch_spec_workspace_execution_jobs (batch_spec_workspace_id, user_id)
    SELECT batch_spec_workspaces.id,
           batch_specs.user_id
    FROM batch_spec_workspaces
             JOIN batch_specs ON batch_specs.id = batch_spec_workspaces.batch_spec_id
    WHERE batch_spec_workspaces.id = ANY (workspace_ids);
END;
$$ LANGUAGE plpgsql;
