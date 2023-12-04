-- drop view
DROP VIEW IF EXISTS gitserver_relocator_jobs_with_repo_name;

-- drop the table
DROP TABLE IF EXISTS gitserver_relocator_jobs;

-- create the old table
CREATE TABLE IF NOT EXISTS gitserver_localclone_jobs (
    id                  SERIAL PRIMARY KEY,
    state               text DEFAULT 'queued',
    queued_at           timestamptz DEFAULT NOW(),
    failure_message     text,
    started_at          timestamp with time zone,
    finished_at         timestamp with time zone,
    process_after       timestamp with time zone,
    num_resets          integer not null DEFAULT 0,
    num_failures        integer not null DEFAULT 0,
    last_heartbeat_at   timestamp with time zone,
    execution_logs      json[],
    worker_hostname     text not null DEFAULT '',

    repo_id             integer not null,
    source_hostname     text not null,
    dest_hostname       text not null,
    delete_source       boolean not null DEFAULT false
);

-- create the old view
CREATE OR REPLACE VIEW gitserver_localclone_jobs_with_repo_name AS SELECT glj.id,
  glj.state,
  glj.failure_message,
  glj.started_at,
  glj.finished_at,
  glj.process_after,
  glj.num_resets,
  glj.num_failures,
  glj.last_heartbeat_at,
  glj.execution_logs,
  glj.worker_hostname,
  glj.repo_id,
  glj.source_hostname,
  glj.dest_hostname,
  glj.delete_source,
  glj.queued_at,
  r.name AS repo_name
FROM (gitserver_localclone_jobs glj
JOIN repo r ON ((r.id = glj.repo_id)));
