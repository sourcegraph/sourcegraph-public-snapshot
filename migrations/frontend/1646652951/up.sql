-- Perform migration here.
--
-- See /migrations/README.md. Highlights:
--  * Make migrations idempotent (use IF EXISTS)
--  * Make migrations backwards-compatible (old readers/writers must continue to work)
--  * If you are using CREATE INDEX CONCURRENTLY, then make sure that only one statement
--    is defined per file, and that each such statement is NOT wrapped in a transaction.
--    Each such migration must also declare "createIndexConcurrently: true" in their
--    associated metadata.yaml file.
CREATE TABLE IF NOT EXISTS orgs_open_beta_stats (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id integer,
    org_id integer,
    created_at timestamptz DEFAULT now(),
    data jsonb NOT NULL DEFAULT '{}'
);
