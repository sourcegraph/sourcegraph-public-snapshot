BEGIN;

UPDATE changesets SET external_deleted_at = NULL WHERE external_deleted_at IS NOT NULL;

COMMIT;
