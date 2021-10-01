CREATE INDEX CONCURRENTLY IF NOT EXISTS security_event_logs_source ON security_event_logs USING btree (source);
