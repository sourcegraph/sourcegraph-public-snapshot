BEGIN;

-- Recreate unique constraint to exclude provider column.
ALTER TABLE user_permissions
    DROP CONSTRAINT IF EXISTS user_permissions_perm_object_provider_unique,
    DROP CONSTRAINT IF EXISTS user_permissions_perm_object_unique,
    ADD CONSTRAINT user_permissions_perm_object_unique
        UNIQUE (user_id, permission, object_type);

ALTER TABLE repo_permissions
    DROP CONSTRAINT IF EXISTS repo_permissions_perm_provider_unique,
    DROP CONSTRAINT IF EXISTS repo_permissions_perm_unique,
    ADD CONSTRAINT repo_permissions_perm_unique
        UNIQUE (repo_id, permission);

ALTER TABLE user_permissions ALTER COLUMN provider DROP NOT NULL;
ALTER TABLE repo_permissions ALTER COLUMN provider DROP NOT NULL;

-- To be done in 3.15:
-- Remove the provider column.
-- ALTER TABLE user_permissions DROP COLUMN provider;
-- ALTER TABLE repo_permissions DROP COLUMN provider;

COMMIT;
