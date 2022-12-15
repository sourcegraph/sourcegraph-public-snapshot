ALTER TABLE executor_secret_access_logs
ADD COLUMN IF NOT EXISTS machine_user text NOT NULL DEFAULT '';

ALTER TABLE executor_secret_access_logs
DROP CONSTRAINT IF EXISTS user_id_or_machine_user;

ALTER TABLE executor_secret_access_logs
ADD CONSTRAINT user_id_or_machine_user
CHECK (
    (user_id IS NULL AND machine_user <> '') OR
    (user_id IS NOT NULL AND machine_user = '')
);

ALTER TABLE executor_secret_access_logs
ALTER COLUMN user_id
DROP NOT NULL;

ALTER TABLE lsif_indexes
ADD COLUMN IF NOT EXISTS requested_envvars text[];

DROP VIEW IF EXISTS lsif_indexes_with_repository_name;

CREATE VIEW lsif_indexes_with_repository_name AS
    SELECT u.id,
        u.commit,
        u.queued_at,
        u.state,
        u.failure_message,
        u.started_at,
        u.finished_at,
        u.repository_id,
        u.process_after,
        u.num_resets,
        u.num_failures,
        u.docker_steps,
        u.root,
        u.indexer,
        u.indexer_args,
        u.outfile,
        u.log_contents,
        u.execution_logs,
        u.local_steps,
        u.should_reindex,
        u.requested_envvars,
        r.name AS repository_name
    FROM (lsif_indexes u
        JOIN repo r ON ((r.id = u.repository_id)))
    WHERE (r.deleted_at IS NULL);
