-- Undo the changes made in the up migration
ALTER TABLE security_event_logs ADD CONSTRAINT security_event_logs_check_has_user
    CHECK (user_id = 0 AND anonymous_user_id <> ''::text OR user_id <> 0 AND anonymous_user_id = ''::text OR user_id <> 0 AND anonymous_user_id <> ''::text);
