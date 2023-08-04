CREATE INDEX IF NOT EXISTS event_logs_name ON event_logs USING GIN (name gin_trgm_ops);
