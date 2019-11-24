BEGIN;

-- Fill in values for rows belong to Bitbucket Server Authz Provider.
ALTER TABLE user_permissions ADD COLUMN provider TEXT;
UPDATE user_permissions SET provider = 'bitbucketServer';
ALTER TABLE user_permissions ALTER COLUMN provider SET NOT NULL;

-- Recreate unique constraint to include provider column.
ALTER TABLE user_permissions
    DROP CONSTRAINT IF EXISTS user_permissions_perm_object_unique,
    DROP CONSTRAINT IF EXISTS user_permissions_perm_object_provider_unique,
    ADD CONSTRAINT user_permissions_perm_object_provider_unique
        UNIQUE (user_id, permission, object_type, provider);

-- Create the inverse table of user_permissions.
CREATE TABLE IF NOT EXISTS repo_permissions (
    repo_id     INTEGER NOT NULL,
    permission  TEXT NOT NULL,
    user_ids    BYTEA NOT NULL,
    provider    TEXT NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);

ALTER TABLE repo_permissions
    DROP CONSTRAINT IF EXISTS repo_permissions_perm_provider_unique,
    ADD CONSTRAINT repo_permissions_perm_provider_unique
        UNIQUE (repo_id, permission, provider);

-- Create the table for pending permissions for users who are not yet able to grant.
-- The bind_id is a text column which represent either username or email, depending on the site config.
-- Example insert:
--     INSERT INTO user_pending_permissions
--       (bind_id, permission, object_type, object_ids, updated_at)
--     VALUES
--       (“joe”, "read", "repos", <bitmap of repo IDs>, NOW());
CREATE TABLE IF NOT EXISTS user_pending_permissions (
    id          SERIAL,
    bind_id     TEXT NOT NULL,
    permission  TEXT NOT NULL,
    object_type TEXT NOT NULL,
    object_ids  BYTEA NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);

ALTER TABLE user_pending_permissions
    DROP CONSTRAINT IF EXISTS user_pending_permissions_perm_object_unique,
    ADD CONSTRAINT user_pending_permissions_perm_object_unique
        UNIQUE (bind_id, permission, object_type);

-- Create the inverse table of user_pending_permissions for performant CRUD.
-- Example insert:
--     INSERT INTO user_pending_permissions
--       (repo_id, permission, user_ids, updated_at)
--     VALUES
--       (1, "read", <bitmap of user IDs>, NOW());
CREATE TABLE IF NOT EXISTS repo_pending_permissions (
    repo_id     INTEGER NOT NULL,
    permission  TEXT NOT NULL,
    user_ids    BYTEA NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);

ALTER TABLE repo_pending_permissions
    DROP CONSTRAINT IF EXISTS repo_pending_permissions_perm_unique,
    ADD CONSTRAINT repo_pending_permissions_perm_unique
        UNIQUE (repo_id, permission);

COMMIT;
