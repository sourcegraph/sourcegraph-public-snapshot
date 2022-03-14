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

CREATE TABLE IF NOT EXISTS cm_last_searched (
    monitor_id BIGINT NOT NULL REFERENCES cm_monitors(id) ON DELETE CASCADE,
    args_hash BIGINT NOT NULL,
    commit_oids text[] NOT NULL,
    PRIMARY KEY (monitor_id, args_hash)
);

COMMENT ON TABLE cm_last_searched
    IS 'The last searched commit hashes for the given code monitor and unique set of search arguments';
COMMENT ON COLUMN cm_last_searched.args_hash
    IS 'A unique hash of the gitserver search arguments to identify this search job';
COMMENT ON COLUMN cm_last_searched.commit_oids
    IS 'The set of commit OIDs that was previously successfully searched and should be excluded on the next run';
