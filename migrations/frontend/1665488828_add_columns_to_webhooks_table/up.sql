ALTER TABLE webhooks
    DROP COLUMN IF EXISTS created_by_user_id,
    DROP COLUMN IF EXISTS updated_by_user_id,
    ADD COLUMN created_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN updated_by_user_id INTEGER DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL;

COMMENT ON COLUMN webhooks.created_by_user_id IS 'ID of a user, who created the webhook. If NULL, then the user does not exist (never existed or was deleted).';
COMMENT ON COLUMN webhooks.updated_by_user_id IS 'ID of a user, who updated the webhook. If NULL, then the user does not exist (never existed or was deleted).';
