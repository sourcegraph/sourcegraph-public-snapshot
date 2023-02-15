CREATE TABLE IF NOT EXISTS permission_sync_jobs (
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

  repository_id integer,
  user_id       integer,

  high_priority     boolean not null default false,
  invalidate_caches boolean not null default false
);

CREATE INDEX IF NOT EXISTS permission_sync_jobs_state ON permission_sync_jobs (state);
CREATE INDEX IF NOT EXISTS permission_sync_jobs_process_after ON permission_sync_jobs (process_after);
CREATE INDEX IF NOT EXISTS permission_sync_jobs_repository_id ON permission_sync_jobs (repository_id);
CREATE INDEX IF NOT EXISTS permission_sync_jobs_user_id ON permission_sync_jobs (user_id);
