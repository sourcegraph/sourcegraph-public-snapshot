CREATE INDEX CONCURRENTLY IF NOT EXISTS security_event_logs_user_id ON security_event_logs USING btree (user_id);
