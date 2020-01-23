BEGIN;

ALTER TABLE event_logs ALTER COLUMN argument DROP NOT NULL;
UPDATE event_logs SET argument = NULL;
ALTER TABLE event_logs ALTER COLUMN argument TYPE jsonb USING argument::jsonb;

COMMIT;
