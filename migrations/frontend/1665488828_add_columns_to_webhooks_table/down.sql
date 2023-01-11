ALTER TABLE webhooks
    DROP COLUMN IF EXISTS created_by_user_id,
    DROP COLUMN IF EXISTS updated_by_user_id;
