BEGIN;

DELETE FROM changeset_events WHERE kind LIKE 'gitlab:%';

COMMIT;
