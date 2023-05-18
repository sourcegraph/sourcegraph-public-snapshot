ALTER TABLE namespace_permissions
    ADD COLUMN IF NOT EXISTS action text NOT NULL;

ALTER TABLE namespace_permissions DROP CONSTRAINT IF EXISTS unique_resource_permission;

CREATE UNIQUE INDEX IF NOT EXISTS unique_resource_permission ON namespace_permissions (namespace, resource_id, action, user_id);

DROP INDEX IF EXISTS unique_resource_permission;
ALTER TABLE namespace_permissions DROP CONSTRAINT IF EXISTS unique_resource_permission;
ALTER TABLE namespace_permissions ADD CONSTRAINT unique_resource_permission UNIQUE (namespace, resource_id, action, user_id);
ALTER TABLE namespace_permissions DROP CONSTRAINT IF EXISTS action_not_blank;
ALTER TABLE namespace_permissions ADD CONSTRAINT action_not_blank CHECK (action <> ''::text);
