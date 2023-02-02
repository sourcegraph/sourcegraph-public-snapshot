ALTER TABLE permission_sync_jobs ADD COLUMN IF NOT EXISTS permissions_added Integer;
ALTER TABLE permission_sync_jobs ADD COLUMN IF NOT EXISTS permissions_removed Integer;
ALTER TABLE permission_sync_jobs ADD COLUMN IF NOT EXISTS permissions_found Integer;
