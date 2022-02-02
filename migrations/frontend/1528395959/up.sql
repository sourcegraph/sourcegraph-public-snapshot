BEGIN;

ALTER TABLE batch_changes RENAME COLUMN initial_applier_id TO creator_id;

COMMIT;
