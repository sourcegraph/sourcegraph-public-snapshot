BEGIN;

ALTER TABLE event_logs DROP CONSTRAINT event_logs_check_url_not_empty;

COMMIT;
