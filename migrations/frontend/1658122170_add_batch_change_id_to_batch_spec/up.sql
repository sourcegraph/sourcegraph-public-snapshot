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

ALTER TABLE batch_specs
    ADD COLUMN IF NOT EXISTS batch_change_id integer,
    ADD FOREIGN KEY (batch_change_id) REFERENCES batch_changes(id) ON DELETE SET NULL DEFERRABLE;

UPDATE batch_specs SET batch_change_id = (SELECT bc.id FROM batch_changes bc WHERE bc.id = batch_specs.id);
