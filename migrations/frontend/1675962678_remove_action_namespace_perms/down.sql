ALTER TABLE namespace_permissions
ADD COLUMN IF NOT EXISTS action text NOT NULL;
