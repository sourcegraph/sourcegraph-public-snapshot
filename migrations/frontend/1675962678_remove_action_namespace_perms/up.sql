ALTER TABLE namespace_permissions DROP COLUMN IF EXISTS action;

ALTER TABLE namespace_permissions DROP CONSTRAINT IF EXISTS unique_resource_permission;

CREATE UNIQUE INDEX IF NOT EXISTS unique_resource_permission ON namespace_permissions (namespace, resource_id, user_id);
