BEGIN;

ALTER TABLE changeset_events DROP COLUMN source;
ALTER TABLE changeset_events ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now();

COMMIT;
