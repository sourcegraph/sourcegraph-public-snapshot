ALTER TABLE namespace_permissions
    ADD COLUMN IF NOT EXISTS action text NOT NULL;

ALTER TABLE namespace_permissions DROP CONSTRAINT IF EXISTS unique_resource_permission;

CREATE UNIQUE INDEX IF NOT EXISTS unique_resource_permission ON namespace_permissions (namespace, resource_id, action, user_id);
