ALTER TABLE permission_sync_jobs
    DROP COLUMN IF EXISTS reason,
    DROP COLUMN IF EXISTS triggered_by_user_id;
