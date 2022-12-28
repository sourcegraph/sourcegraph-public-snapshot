CREATE TABLE IF NOT EXISTS outbound_webhook_jobs (
    id BIGSERIAL NOT NULL PRIMARY KEY,

    -- Requirements to send a payload.
    event_type TEXT NOT NULL,
    scope TEXT NULL,
    encryption_key_id TEXT NULL,
    payload BYTEA NOT NULL,

    -- Generic worker fields.
    state TEXT NOT NULL DEFAULT 'queued',
    failure_message TEXT,
    queued_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    process_after TIMESTAMP WITH TIME ZONE,
    num_resets INTEGER NOT NULL DEFAULT 0,
    num_failures INTEGER NOT NULL DEFAULT 0,
    last_heartbeat_at TIMESTAMP WITH TIME ZONE NULL,
    execution_logs JSON[],
    worker_hostname TEXT NOT NULL DEFAULT '',
    cancel BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS outbound_webhook_jobs_state_idx
ON outbound_webhook_jobs (state);

CREATE INDEX IF NOT EXISTS outbound_webhook_payload_process_after_idx
ON outbound_webhook_jobs (process_after);

CREATE TABLE IF NOT EXISTS outbound_webhook_logs (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    job_id BIGINT NOT NULL REFERENCES outbound_webhook_jobs (id) ON DELETE CASCADE ON UPDATE CASCADE,
    outbound_webhook_id BIGINT NOT NULL REFERENCES outbound_webhooks (id) ON DELETE CASCADE ON UPDATE CASCADE,
    sent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    status_code INTEGER NOT NULL,
    encryption_key_id TEXT NULL,
    request BYTEA NOT NULL,
    response BYTEA NOT NULL,
    error BYTEA NOT NULL
);

CREATE INDEX IF NOT EXISTS outbound_webhook_logs_outbound_webhook_id_idx
ON outbound_webhook_logs (outbound_webhook_id);

CREATE INDEX IF NOT EXISTS outbound_webhooks_logs_status_code_idx
ON outbound_webhook_logs (status_code);
