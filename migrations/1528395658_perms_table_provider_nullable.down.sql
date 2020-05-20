BEGIN;

-- Fill in values for rows as Bitbucket Server authz provider,
-- no customers known have used Sourcegraph authz provider yet.
UPDATE user_permissions SET provider = 'bitbucketServer';
ALTER TABLE user_permissions ALTER COLUMN provider SET NOT NULL;

UPDATE repo_permissions SET provider = 'bitbucketServer';
ALTER TABLE repo_permissions ALTER COLUMN provider SET NOT NULL;

-- Recreate unique constraint to include provider column.
ALTER TABLE user_permissions
    DROP CONSTRAINT IF EXISTS user_permissions_perm_object_unique,
    DROP CONSTRAINT IF EXISTS user_permissions_perm_object_provider_unique,
    ADD CONSTRAINT user_permissions_perm_object_provider_unique
        UNIQUE (user_id, permission, object_type, provider);

ALTER TABLE repo_permissions
    DROP CONSTRAINT IF EXISTS repo_permissions_perm_unique,
    DROP CONSTRAINT IF EXISTS repo_permissions_perm_provider_unique,
    ADD CONSTRAINT repo_permissions_perm_provider_unique
        UNIQUE (repo_id, permission, provider);

COMMIT;
