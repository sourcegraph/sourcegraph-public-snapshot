BEGIN;

-- Restore old unique constraint
ALTER TABLE user_permissions
    DROP CONSTRAINT IF EXISTS user_permissions_perm_object_provider_unique,
    DROP CONSTRAINT IF EXISTS user_permissions_perm_object_unique,
    ADD CONSTRAINT user_permissions_perm_object_unique
        UNIQUE (user_id, permission, object_type);
ALTER TABLE user_permissions DROP COLUMN IF EXISTS provider;

-- Drop new tables created
DROP TABLE IF EXISTS repo_permissions;
DROP TABLE IF EXISTS user_pending_permissions;
DROP TABLE IF EXISTS repo_pending_permissions;

COMMIT;
