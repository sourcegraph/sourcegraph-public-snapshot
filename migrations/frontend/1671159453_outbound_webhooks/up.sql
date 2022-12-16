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

CREATE TABLE IF NOT EXISTS outbound_webhook_event_types (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    outbound_webhook_id BIGINT NOT NULL REFERENCES outbound_webhooks (id) ON DELETE CASCADE ON UPDATE CASCADE,
    event_type TEXT NOT NULL,
    scope TEXT NULL
);

CREATE INDEX IF NOT EXISTS outbound_webhook_event_types_event_type_idx
ON outbound_webhook_event_types (event_type, scope);

DROP VIEW IF EXISTS outbound_webhooks_with_event_types;

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
