ALTER TABLE webhooks
    DROP COLUMN IF EXISTS created_by,
    DROP COLUMN IF EXISTS updated_by,
    ADD COLUMN created_by INTEGER,
    ADD COLUMN updated_by INTEGER;

COMMENT ON COLUMN webhooks.created_by IS 'ID of a user, who created the webhook. If equals to zero, then the webhook is updated programmatically.';
COMMENT ON COLUMN webhooks.updated_by IS 'ID of a user, who updated the webhook. If equals to zero, then the webhook is updated programmatically.';
