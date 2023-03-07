-- First set migrated column to false on all rows
ALTER TABLE IF EXISTS user_permissions
    ADD COLUMN IF NOT EXISTS migrated BOOLEAN DEFAULT FALSE;
-- Now default migrated to true, e.g. all upserted rows from now on are automatically migrated
-- since we are writing to both user_repo_permissions and user_permissions tables
ALTER TABLE IF EXISTS user_permissions
    ALTER COLUMN migrated SET DEFAULT TRUE;
