CREATE TABLE IF NOT EXISTS last_lockfile_scan (
    repository_id integer NOT NULL PRIMARY KEY,
    last_lockfile_scan_at timestamp with time zone NOT NULL
);

COMMENT ON TABLE last_lockfile_scan IS 'Tracks the last time repository was checked for lockfile indexing.';

COMMENT ON COLUMN last_lockfile_scan.last_lockfile_scan_at IS 'The last time this repository was considered for lockfile indexing.';
