BEGIN;

UPDATE event_logs SET argument = '{}';
ALTER TABLE event_logs ALTER COLUMN argument TYPE jsonb USING argument::jsonb;

COMMIT;
