BEGIN;

-- We do not recreate the grants, as we've shifted our strategy away from row-
-- level security to application-level code. Prior migrations that created the
-- grants have also been removed.

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
