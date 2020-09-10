BEGIN;

CREATE INDEX IF NOT EXISTS user_external_accounts_user_id
    ON user_external_accounts USING BTREE (user_id)
    WHERE deleted_at IS NULL;

COMMIT;
