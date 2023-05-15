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

CREATE UNIQUE INDEX own_signal_configurations_name_uidx ON own_signal_configurations(name);

INSERT INTO own_signal_configurations (id, name, enabled, description)
VALUES (1, 'recent-contributors', FALSE, 'Indexes contributors in each file using repository history.')
ON CONFLICT DO NOTHING;
INSERT INTO own_signal_configurations (id, name, enabled, description)
VALUES (2, 'recent-views', FALSE, 'Indexes users that recently viewed files in Sourcegraph.')
ON CONFLICT DO NOTHING;

