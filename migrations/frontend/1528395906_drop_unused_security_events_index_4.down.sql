CREATE INDEX CONCURRENTLY IF NOT EXISTS security_event_logs_timestamp_at_utc ON security_event_logs USING btree (date(timezone('UTC'::text, "timestamp")));
