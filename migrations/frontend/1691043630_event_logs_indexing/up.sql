CREATE INDEX IF NOT EXISTS event_logs_name ON event_logs (name);

CREATE INDEX IF NOT EXISTS event_logs_name_lower ON event_logs (lower(name));

CREATE INDEX IF NOT EXISTS event_logs_timestamp ON event_logs USING btree ("timestamp");
