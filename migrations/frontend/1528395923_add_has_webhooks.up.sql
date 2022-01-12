BEGIN;

ALTER TABLE
    external_services
ADD COLUMN IF NOT EXISTS
    has_webhooks BOOLEAN NULL DEFAULT NULL;

CREATE INDEX
    external_services_has_webhooks_idx
ON
    external_services (has_webhooks);

INSERT INTO
    out_of_band_migrations (
        id,
        team,
        component,
        description,
        introduced_version_major,
        introduced_version_minor,
        non_destructive
    )
VALUES (
    13,
    'batch-changes',
    'frontend-db.external_services',
    'Calculate the webhook state of each external service',
    3,
    34,
    true
)
ON CONFLICT
    DO NOTHING
;

COMMIT;
