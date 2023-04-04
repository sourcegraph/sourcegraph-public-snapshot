-- Create the basic outbound webhook registry table.
CREATE TABLE IF NOT EXISTS outbound_webhooks (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    created_by INTEGER NULL REFERENCES users (id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_by INTEGER NULL REFERENCES users (id) ON DELETE SET NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    encryption_key_id TEXT NULL,
    url BYTEA NOT NULL,
    secret BYTEA NOT NULL
);

-- Each webhook can have one or many event types associated with it, so we'll
-- use a separate table to track that.
CREATE TABLE IF NOT EXISTS outbound_webhook_event_types (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    outbound_webhook_id BIGINT NOT NULL REFERENCES outbound_webhooks (id) ON DELETE CASCADE ON UPDATE CASCADE,
    event_type TEXT NOT NULL,
    scope TEXT NULL
);

CREATE INDEX IF NOT EXISTS outbound_webhook_event_types_event_type_idx
ON outbound_webhook_event_types (event_type, scope);

DROP VIEW IF EXISTS outbound_webhooks_with_event_types;

-- Most of the time we interact with webhooks from code, we want to also hydrate
-- the event types. This view does exactly that.
CREATE VIEW outbound_webhooks_with_event_types AS
SELECT
    id,
    created_by,
    created_at,
    updated_by,
    updated_at,
    encryption_key_id,
    url,
    secret,
    array_to_json(
        array(
            SELECT
                json_build_object(
                    'id', id,
                    'outbound_webhook_id', outbound_webhook_id,
                    'event_type', event_type,
                    'scope', scope
                )
            FROM
                outbound_webhook_event_types
            WHERE
                outbound_webhook_id = outbound_webhooks.id
        )
     ) AS event_types
FROM
    outbound_webhooks;

-- Create a table to track outbound webhook worker jobs.
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

-- Create a table to store outbound webhook logs.
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
