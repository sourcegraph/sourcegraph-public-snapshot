CREATE INDEX IF NOT EXISTS event_logs_name_is_cody_explanation_event ON event_logs USING btree (iscodyexplanationevent(name));
