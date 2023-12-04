ALTER TABLE IF EXISTS event_logs
    ADD COLUMN IF NOT EXISTS first_source_url TEXT,
    ADD COLUMN IF NOT EXISTS last_source_url  TEXT,
    ADD COLUMN IF NOT EXISTS referrer         TEXT,
    ADD COLUMN IF NOT EXISTS device_id        TEXT,
    ADD COLUMN IF NOT EXISTS insert_id        TEXT;
