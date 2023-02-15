-- ALTER TABLE IF EXISTS
--     roles
-- RENAME COLUMN system TO readonly;

-- The above query isn't idempotent because `RENAME COLUMN` doesn't have an `IF EXISTS`
-- clause that checks for the existence of the column in question.

DO $$
BEGIN
    ALTER TABLE IF EXISTS
        roles
    RENAME COLUMN system TO readonly;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column system does not exist in table roles';
END $$;
