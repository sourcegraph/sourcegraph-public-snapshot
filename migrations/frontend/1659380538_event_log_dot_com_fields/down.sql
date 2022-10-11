ALTER TABLE IF EXISTS event_logs
    DROP COLUMN IF EXISTS first_source_url,
    DROP COLUMN IF EXISTS last_source_url,
    DROP COLUMN IF EXISTS referrer,
    DROP COLUMN IF EXISTS device_id,
    DROP COLUMN IF EXISTS insert_id;
