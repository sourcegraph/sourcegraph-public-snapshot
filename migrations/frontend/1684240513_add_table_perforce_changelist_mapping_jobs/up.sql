CREATE TABLE IF NOT EXISTS perforce_changelist_mapping_jobs (
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

  repo_id integer not null
);

CREATE UNIQUE INDEX IF NOT EXISTS perforce_changelist_mapping_jobs_id_repo_id_unique ON perforce_changelist_mapping_jobs USING btree (id, repo_id);
CREATE INDEX IF NOT EXISTS perforce_changelist_mapping_jobs_state ON perforce_changelist_mapping_jobs USING btree(state);
CREATE INDEX IF NOT EXISTS perforce_changelist_mapping_jobs_process_after ON perforce_changelist_mapping_jobs USING btree(process_after);
