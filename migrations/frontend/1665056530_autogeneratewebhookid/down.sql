ALTER TABLE IF EXISTS webhooks ALTER COLUMN id DROP DEFAULT;

ALTER TABLE IF EXISTS webhooks
    DROP CONSTRAINT IF EXISTS webhooks_pkey;
