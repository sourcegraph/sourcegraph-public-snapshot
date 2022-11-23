ALTER TABLE webhooks
    DROP COLUMN IF EXISTS name,
    ADD COLUMN name TEXT;

COMMENT ON COLUMN webhooks.name IS 'Descriptive name of a webhook.';

UPDATE webhooks SET name = code_host_urn WHERE name IS NULL;

ALTER TABLE webhooks
    ALTER COLUMN name SET NOT NULL;
