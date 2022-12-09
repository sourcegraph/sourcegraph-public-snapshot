-- https://docs.sourcegraph.com/dev/background-information/workers#step-1-create-a-jobs-table

CREATE TABLE IF NOT EXISTS own_blame_jobs (
  id                SERIAL PRIMARY KEY,
  state             text DEFAULT 'queued',
  failure_message   text,
  queued_at         timestamp with time zone DEFAULT NOW(),
  started_at        timestamp with time zone,
  finished_at       timestamp with time zone,
  process_after     timestamp with time zone,
  num_resets        integer not null default 0,
  num_failures      integer not null default 0,
  last_heartbeat_at timestamp with time zone,
  execution_logs    json[],
  worker_hostname   text not null default '',
  cancel            boolean not null default false,

  repository_id integer not null,
  absolute_file_path text not null
);

CREATE OR REPLACE VIEW own_blame_jobs_with_repository_name AS
  SELECT j.*, r.name
  FROM own_blame_jobs j
  JOIN repo r ON r.id = j.repository_id;
