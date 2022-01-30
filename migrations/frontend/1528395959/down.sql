BEGIN;

ALTER TABLE batch_changes RENAME COLUMN creator_id TO initial_applier_id;

COMMIT;
