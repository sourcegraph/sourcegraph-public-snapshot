CREATE TABLE IF NOT EXISTS explicit_permissions_bitbucket_projects_jobs (
    id SERIAL PRIMARY KEY,
    state text DEFAULT 'queued',
    failure_message text,
    queued_at timestamp with time zone DEFAULT NOW(),
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer not null default 0,
    num_failures integer not null default 0,
    last_heartbeat_at timestamp with time zone,
    execution_logs json [],
    worker_hostname text not null default '',
    -- additional columns
    project_key text not null,
    external_service_id integer not null,
    permissions json [],
    unrestricted boolean not null default false,
    CHECK (
        (
            permissions IS NOT NULL
            AND unrestricted IS false
        )
        OR (
            permissions IS NULL
            AND unrestricted IS true
        )
    )
);
