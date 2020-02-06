BEGIN;

DELETE FROM changeset_events
WHERE kind = 'github:labeled' OR kind = 'github:unlabeled';

COMMIT;
