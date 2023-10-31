ALTER TABLE event_logs DROP CONSTRAINT event_logs_check_has_user;

ALTER TABLE event_logs ADD CONSTRAINT event_logs_check_has_user
    CHECK (
        user_id = 0 AND anonymous_user_id <> ''::text OR
        user_id <> 0 AND anonymous_user_id = ''::text OR
        user_id <> 0 AND anonymous_user_id <> ''::text
        );
