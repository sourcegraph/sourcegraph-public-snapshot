CREATE INDEX IF NOT EXISTS event_logs_name ON event_logs (name);

CREATE INDEX IF NOT EXISTS event_logs_timestamp ON event_logs USING btree ("timestamp");
