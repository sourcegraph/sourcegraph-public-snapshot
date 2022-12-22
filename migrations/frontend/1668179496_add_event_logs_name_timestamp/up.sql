CREATE INDEX CONCURRENTLY IF NOT EXISTS event_logs_name_timestamp ON event_logs(name, timestamp desc);
