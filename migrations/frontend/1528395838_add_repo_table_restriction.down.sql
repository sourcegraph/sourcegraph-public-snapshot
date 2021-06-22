BEGIN;

-- This unwinds the row-level security additions to the repo table
-- and removes the "sgservice" role from the system.
--
-- The "sgservice" role does not own anything, so it's safe to drop.
-- This will remove all its privileges, which will allow the role to
-- be removed.
DROP OWNED  BY sgservice;
DROP POLICY IF EXISTS restricted_repo_policy ON repo;
DROP ROLE   IF EXISTS sgservice;

-- With row-level security disabled, all roles will be able to access
-- all rows in the table.
ALTER TABLE repo DISABLE ROW LEVEL SECURITY;

COMMIT;
