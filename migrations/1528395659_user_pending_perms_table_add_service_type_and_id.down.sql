BEGIN;

-- Roll back the constraint
ALTER TABLE user_pending_permissions
    DROP CONSTRAINT IF EXISTS user_pending_permissions_perm_object_service_unique,
    DROP CONSTRAINT IF EXISTS user_pending_permissions_perm_object_unique,
    ADD CONSTRAINT user_pending_permissions_perm_object_unique
        UNIQUE (bind_id, permission, object_type);

ALTER TABLE user_pending_permissions DROP COLUMN IF EXISTS service_type;
ALTER TABLE user_pending_permissions DROP COLUMN IF EXISTS service_id;

COMMIT;
