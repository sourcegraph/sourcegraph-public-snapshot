-- drop view
DROP VIEW IF EXISTS gitserver_localclone_jobs_with_repo_name;

-- drop the table
DROP TABLE IF EXISTS gitserver_localclone_jobs;

-- create the new table
CREATE TABLE IF NOT EXISTS gitserver_relocator_jobs (
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

-- create the view
CREATE OR REPLACE VIEW gitserver_relocator_jobs_with_repo_name AS
  SELECT glj.*, r.name AS repo_name
  FROM gitserver_relocator_jobs glj
  JOIN repo r ON r.id = glj.repo_id;
