BEGIN;

ALTER TABLE changeset_events DROP COLUMN updated_at;
ALTER TABLE changeset_events ADD COLUMN source TEXT NOT NULL CHECK (source != '');

COMMIT;
