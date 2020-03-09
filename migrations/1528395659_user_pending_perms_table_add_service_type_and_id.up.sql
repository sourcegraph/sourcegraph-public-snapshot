BEGIN;

-- Add new columns to the pending permissions tables.
-- The Sourcegraph authz provider is the only one uses these tables,
-- thus safe to set default values for existing rows.
ALTER TABLE user_pending_permissions ADD COLUMN service_type TEXT;
UPDATE user_pending_permissions SET service_type = 'sourcegraph';
ALTER TABLE user_pending_permissions ALTER COLUMN service_type SET NOT NULL;

ALTER TABLE user_pending_permissions ADD COLUMN service_id TEXT;
UPDATE user_pending_permissions SET service_id = 'https://sourcegraph.com/';
ALTER TABLE user_pending_permissions ALTER COLUMN service_id SET NOT NULL;

-- Recreate unique constraint to include new columns.

-- Example insert:
--     INSERT INTO user_pending_permissions
--       (bind_id, permission, object_type, object_ids, service_type, service_id, updated_at)
--     VALUES
--       ("joe", "read", "repos", <bitmap of repo IDs>, "gitlab", "https://gitlab.com/", NOW());
ALTER TABLE user_pending_permissions
    DROP CONSTRAINT IF EXISTS user_pending_permissions_perm_object_unique,
    DROP CONSTRAINT IF EXISTS user_pending_permissions_service_perm_object_unique,
    ADD CONSTRAINT user_pending_permissions_service_perm_object_unique
        UNIQUE (service_type, service_id, permission, object_type, bind_id);

COMMIT;
