ALTER TABLE webhooks
    ADD COLUMN IF NOT EXISTS name TEXT;

COMMENT ON COLUMN webhooks.name IS 'Descriptive name of a webhook.';

UPDATE webhooks SET name = code_host_urn WHERE name IS NULL;

ALTER TABLE webhooks
    ALTER COLUMN name SET NOT NULL;
