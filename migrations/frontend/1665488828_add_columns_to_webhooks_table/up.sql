ALTER TABLE webhooks
    DROP COLUMN IF EXISTS created_by_user_id,
    DROP COLUMN IF EXISTS updated_by_user_id,
    ADD COLUMN created_by_user_id INTEGER,
    ADD COLUMN updated_by_user_id INTEGER;

COMMENT ON COLUMN webhooks.created_by_user_id IS 'ID of a user, who created the webhook. If equals to zero, then the webhook is updated programmatically.';
COMMENT ON COLUMN webhooks.updated_by_user_id IS 'ID of a user, who updated the webhook. If equals to zero, then the webhook is updated programmatically.';
