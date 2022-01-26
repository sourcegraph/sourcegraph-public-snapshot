BEGIN;

-- We do not recreate the policy, as we've shifted our strategy away from row-
-- level security to application-level code. Prior migrations that created the
-- policy have also been removed.

COMMIT;
