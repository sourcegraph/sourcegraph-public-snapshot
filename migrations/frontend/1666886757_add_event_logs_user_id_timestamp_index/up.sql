CREATE INDEX CONCURRENTLY IF NOT EXISTS event_logs_user_id_timestamp ON event_logs(user_id, timestamp);
