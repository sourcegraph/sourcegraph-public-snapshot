-- ALTER TABLE IF EXISTS
--     roles
-- RENAME COLUMN readonly TO system;

-- The above query isn't idempotent because `RENAME COLUMN` doesn't have an `IF EXISTS`
-- clause that checks for the existence of the column in question.

DO $$
BEGIN
    ALTER TABLE IF EXISTS
        roles
    RENAME COLUMN readonly TO system;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column readonly does not exist in table roles';
END $$;
