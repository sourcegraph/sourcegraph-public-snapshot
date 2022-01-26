-- +++
-- parent: 1528395916
-- +++

BEGIN;

-- Create table to rate limit indexing scans of a repository
CREATE TABLE lsif_last_index_scan (
    repository_id int NOT NULL,
    last_index_scan_at timestamp with time zone NOT NULL,

    PRIMARY KEY(repository_id)
);
COMMENT ON TABLE lsif_last_index_scan IS 'Tracks the last time repository was checked for auto-indexing job scheduling.';
COMMENT ON COLUMN lsif_last_index_scan.last_index_scan_at IS 'The last time uploads of this repository were considered for auto-indexing job scheduling.';

COMMIT;
