BEGIN;

UPDATE changesets SET publication_state = 'UNPUBLISHED' WHERE unsynced = true;
ALTER TABLE changesets DROP COLUMN IF EXISTS unsynced;

COMMIT;
