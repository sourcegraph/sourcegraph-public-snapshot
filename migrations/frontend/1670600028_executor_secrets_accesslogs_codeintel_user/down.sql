-- Undo the changes made in the up migration
ALTER TABLE executor_secret_access_logs
DROP CONSTRAINT user_id_or_machine_user;

-- delete access logs where machine_user is true

ALTER TABLE executor_secret_access_logs
DROP COLUMN machine_user;


ALTER TABLE executor_secret_access_logs
ALTER COLUMN user_id
SET NOT NULL;

ALTER TABLE lsif_indexes
DROP COLUMN requested_envvars;

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
