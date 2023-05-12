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

CREATE TABLE IF NOT EXISTS own_signal_configurations
(
    id                     SERIAL PRIMARY KEY,
    name                   TEXT    NOT NULL,
    description            TEXT    NOT NULL DEFAULT '',
    excluded_repo_patterns TEXT[]  NULL,
    enabled                BOOLEAN NOT NULL DEFAULT FALSE
);

INSERT INTO own_signal_configurations (id, name, enabled)
VALUES (1, 'recent-contributors', FALSE);
INSERT INTO own_signal_configurations (id, name, enabled)
VALUES (2, 'recent-views', FALSE);

