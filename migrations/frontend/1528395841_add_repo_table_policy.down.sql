BEGIN;

-- With row-level security disabled, all rows in the repo table
-- will be accessible/visible to all roles.
DROP POLICY IF EXISTS sg_repo_access_policy ON repo;
ALTER TABLE repo DISABLE ROW LEVEL SECURITY;

COMMIT;
