ALTER TABLE webhooks
    ADD COLUMN IF NOT EXISTS name TEXT;

COMMENT ON COLUMN webhooks.name IS 'Descriptive name of a webhook.';
