ALTER TABLE permission_sync_jobs ADD COLUMN IF NOT EXISTS permissions_added integer NOT NULL DEFAULT 0;
ALTER TABLE permission_sync_jobs ADD COLUMN IF NOT EXISTS permissions_removed integer NOT NULL DEFAULT 0;
ALTER TABLE permission_sync_jobs ADD COLUMN IF NOT EXISTS permissions_found integer NOT NULL DEFAULT 0;
