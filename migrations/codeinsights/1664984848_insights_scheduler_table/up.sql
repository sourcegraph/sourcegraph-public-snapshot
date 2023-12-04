CREATE TABLE IF NOT EXISTS insights_background_jobs (
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
                              cancel            boolean not null default false
);

CREATE INDEX IF NOT EXISTS insights_jobs_state_idx ON insights_background_jobs(state);
