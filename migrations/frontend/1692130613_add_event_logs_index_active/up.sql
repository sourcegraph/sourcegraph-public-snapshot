CREATE INDEX CONCURRENTLY IF NOT EXISTS event_logs_name_is_cody_active_event ON event_logs USING btree (iscodyactiveevent(name));
