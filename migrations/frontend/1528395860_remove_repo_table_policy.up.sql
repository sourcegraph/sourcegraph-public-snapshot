BEGIN;

-- This removes the row-level security policy (if present), and disables RLS on
-- the repo table. Both operations are idempotent.
DROP POLICY IF EXISTS sg_repo_access_policy ON repo;
ALTER TABLE repo DISABLE ROW LEVEL SECURITY;

COMMIT;
