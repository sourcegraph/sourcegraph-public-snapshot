-- +++
-- parent: 1000000025
-- +++

BEGIN;

-- Perform migration here.
--
-- See /migrations/README.md. Highlights:
--  * Make migrations idempotent (use IF EXISTS)
--  * Make migrations backwards-compatible (old readers/writers must continue to work)
--  * Wrap your changes in a transaction. Note that CREATE INDEX CONCURRENTLY is an exception
--    and cannot be performed in a transaction. For such migrations, ensure that only one
--    statement is defined per migration to prevent query transactions from starting implicitly.

COMMIT;
