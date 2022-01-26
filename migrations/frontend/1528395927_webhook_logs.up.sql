-- +++
-- parent: 1528395926
-- +++

BEGIN;

CREATE TABLE IF NOT EXISTS webhook_logs (
    id BIGSERIAL PRIMARY KEY,
    received_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    external_service_id INTEGER NULL REFERENCES external_services (id) ON DELETE CASCADE ON UPDATE CASCADE,
    status_code INTEGER NOT NULL,
    request BYTEA NOT NULL,
    response BYTEA NOT NULL,
    encryption_key_id TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS
    webhook_logs_received_at_idx
ON
    webhook_logs (received_at);

CREATE INDEX IF NOT EXISTS
    webhook_logs_external_service_id_idx
ON
    webhook_logs (external_service_id);

CREATE INDEX IF NOT EXISTS
    webhook_logs_status_code_idx
ON
    webhook_logs (status_code);

COMMIT;
