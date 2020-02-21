BEGIN;

ALTER TABLE event_logs ADD CONSTRAINT event_logs_check_url_not_empty CHECK ((url <> ''::text));

COMMIT;
