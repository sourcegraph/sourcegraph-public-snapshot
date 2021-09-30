CREATE INDEX CONCURRENTLY IF NOT EXISTS security_event_logs_anonymous_user_id ON security_event_logs USING btree (anonymous_user_id);
