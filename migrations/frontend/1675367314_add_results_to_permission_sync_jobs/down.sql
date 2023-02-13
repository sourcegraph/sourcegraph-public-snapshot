ALTER TABLE permission_sync_jobs DROP COLUMN IF EXISTS permissions_added;
ALTER TABLE permission_sync_jobs DROP COLUMN IF EXISTS permissions_removed;
ALTER TABLE permission_sync_jobs DROP COLUMN IF EXISTS permissions_found;
