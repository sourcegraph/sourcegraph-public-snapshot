DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE
            table_name = 'executor_secret_access_logs_machineuser_bak_1670600028' AND
            table_schema = current_schema()
    ) THEN
        -- create and copy
        CREATE TABLE executor_secret_access_logs_machineuser_bak_1670600028 AS
        SELECT * FROM executor_secret_access_logs
        -- these two should match the same rows, but just in case
        WHERE machine_user <> '' OR user_id IS NULL;
    ELSEIF EXISTS (
        -- must check for double-down idempotency test
        SELECT 1 FROM information_schema.columns
        WHERE
            table_name = 'executor_secret_access_logs' AND
            table_schema = current_schema() AND
            column_name = 'machine_user'
    ) THEN
        -- copy over any rows that may have been added since (unlikely edge-case)
        INSERT INTO executor_secret_access_logs_machineuser_bak_1670600028
        SELECT * FROM executor_secret_access_logs AS esal
        LEFT JOIN executor_secret_access_logs_machineuser_bak_1670600028 AS bak
        ON esal.id = bak.id
        WHERE bak.id IS NULL AND esal.machine_user <> '' OR esal.user_id IS NULL;
    END IF;
END
$$;

ALTER TABLE executor_secret_access_logs
DROP CONSTRAINT IF EXISTS user_id_or_machine_user;

ALTER TABLE executor_secret_access_logs
DROP COLUMN IF EXISTS machine_user;

DELETE FROM executor_secret_access_logs WHERE user_id IS NULL;

ALTER TABLE executor_secret_access_logs
ALTER COLUMN user_id
SET NOT NULL;

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
        r.name AS repository_name
    FROM (lsif_indexes u
        JOIN repo r ON ((r.id = u.repository_id)))
    WHERE (r.deleted_at IS NULL);

ALTER TABLE lsif_indexes
DROP COLUMN IF EXISTS requested_envvars;
