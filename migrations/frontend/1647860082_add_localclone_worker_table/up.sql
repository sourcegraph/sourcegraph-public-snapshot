CREATE TABLE IF NOT EXISTS gitserver_localclone_jobs (
    id                  SERIAL PRIMARY KEY,
    state               text DEFAULT 'queued',
    failure_message     text,
    started_at          timestamp with time zone,
    finished_at         timestamp with time zone,
    process_after       timestamp with time zone,
    num_resets          integer not null default 0,
    num_failures        integer not null default 0,
    last_heartbeat_at   timestamp with time zone,
    execution_logs      json[],
    worker_hostname     text not null default '',

    repo_id             integer not null,
    source_hostname     text not null,
    dest_hostname       text not null,
    delete_source       boolean not null default false
);
