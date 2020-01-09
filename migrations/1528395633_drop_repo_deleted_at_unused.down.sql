BEGIN;

ALTER TABLE repo
      ADD CONSTRAINT deleted_at_unused CHECK ((deleted_at IS NULL));

COMMIT;
