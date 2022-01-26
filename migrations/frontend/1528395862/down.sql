BEGIN;

-- We do not recreate the role, as we've shifted our strategy away from row-
-- level security to application-level code. Prior migrations that created the
-- role have also been removed.

COMMIT;
