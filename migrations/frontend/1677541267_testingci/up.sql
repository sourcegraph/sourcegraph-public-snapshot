-- Perform migration here.
--
-- See /migrations/README.md. Highlights:
--  * Make migrations idempotent (use IF EXISTS)
--  * Make migrations backwards-compatible (old readers/writers must continue to work)
--  * If you are using CREATE INDEX CONCURRENTLY, then make sure that only one statement
--    is defined per file, and that each such statement is NOT wrapped in a transaction.
--    Each such migration must also declare "createIndexConcurrently: true" in their
--    associated metadata.yaml file.
--  * If you are modifying Postgres extensions, you must also declare "privileged: true"
--    in the associated metadata.yaml file.
CREATE TABLE IF NOT EXISTS testci(
    id UUID NOT NULL,
    test TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
    );

COMMENT ON TABLE testci IS 'This is just a test to see if CI would catch migraion';
COMMENT ON COLUMN testci.test IS 'just for test';
