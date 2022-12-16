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

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE
            table_name = 'executor_secret_access_logs_machineuser_bak_1670600028' AND
            table_schema = current_schema()
    ) THEN
        DECLARE
            err_details text;
        BEGIN
            INSERT INTO executor_secret_access_logs
            SELECT * FROM executor_secret_access_logs_machineuser_bak_1670600028;
        EXCEPTION WHEN unique_violation THEN
            GET STACKED DIAGNOSTICS err_details = PG_EXCEPTION_DETAIL;
            RAISE unique_violation
                USING MESSAGE = SQLERRM,
                DETAIL = err_details,
                HINT = 'This up-migration attempts to re-insert executor secret access logs '
                'from a backup table that would have lost information due to the associated '
                'down-migration changing the table schema. In doing so, a unique violation exception '
                'occurred and will have to be resolved manually. The backed up access logs are stored '
                'in the executor_secret_access_logs_machineuser_bak_1670600028 table.';
        END;
    END IF;
END
$$;

DROP TABLE IF EXISTS executor_secret_access_logs_machineuser_bak_1670600028;

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
