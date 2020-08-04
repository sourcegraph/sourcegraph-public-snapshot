BEGIN;

ALTER TABLE external_services ADD COLUMN last_sync_at timestamp with time zone;
ALTER TABLE external_services ADD COLUMN next_sync_at timestamp with time zone;

COMMIT;
